<template>
  <template v-if="pipeline && repo">
    <Scaffold
      v-model:activeTab="activeTab"
      enable-tabs
      disable-hash-mode
      :go-back="goBack"
      :fluid-content="activeTab !== 'tasks'"
    >
      <template #title>
        <span class="w-full md:w-auto text-center">{{ $t('repo.pipeline.pipeline', { pipelineId }) }}</span>
        <span class="<md:hidden">-</span>
        <span class="w-full md:w-auto text-center truncate">{{ message }}</span>
      </template>

      <template #titleActions>
        <PipelineStatusIcon :status="pipeline.status" class="flex flex-shrink-0" />

        <template v-if="repoPermissions.push">
          <Button
            v-if="pipeline.status === 'pending' || pipeline.status === 'running'"
            class="flex-shrink-0"
            :text="$t('repo.pipeline.actions.cancel')"
            :is-loading="isCancelingPipeline"
            @click="cancelPipeline"
          />
          <Button
            v-else-if="pipeline.status !== 'blocked' && pipeline.status !== 'declined'"
            class="flex-shrink-0"
            :text="$t('repo.pipeline.actions.restart')"
            :is-loading="isRestartingPipeline"
            @click="restartPipeline"
          />
        </template>
      </template>

      <template #tabActions>
        <div class="flex justify-between gap-x-4 text-color flex-shrink-0 pb-2 md:p-0 mx-auto md:mr-0">
          <div class="flex space-x-1 items-center flex-shrink-0">
            <Icon name="since" />
            <Tooltip>
              <span>{{ since }}</span>
              <template #popper
                ><span class="font-bold">{{ $t('repo.pipeline.created') }}</span> {{ created }}</template
              >
            </Tooltip>
          </div>
          <div class="flex space-x-1 items-center flex-shrink-0">
            <Icon name="duration" />
            <span>{{ duration }}</span>
          </div>
        </div>
      </template>

      <Tab id="tasks" :title="$t('repo.pipeline.tasks')" />
      <Tab id="config" :title="$t('repo.pipeline.config')" />
      <Tab
        v-if="pipeline.event === 'push' || pipeline.event === 'pull_request'"
        id="changed-files"
        :title="$t('repo.pipeline.files', { files: pipeline.changed_files?.length || 0 })"
      />
      <router-view />
    </Scaffold>
  </template>
</template>

<script lang="ts">
import { Tooltip } from 'floating-vue';
import { computed, defineComponent, inject, onBeforeUnmount, onMounted, provide, Ref, toRef, watch } from 'vue';
import { useI18n } from 'vue-i18n';
import { useRoute, useRouter } from 'vue-router';

import Button from '~/components/atomic/Button.vue';
import Scaffold from '~/components/layout/scaffold/Scaffold.vue';
import Tab from '~/components/layout/scaffold/Tab.vue';
import PipelineStatusIcon from '~/components/repo/pipeline/PipelineStatusIcon.vue';
import useApiClient from '~/compositions/useApiClient';
import { useAsyncAction } from '~/compositions/useAsyncAction';
import { useFavicon } from '~/compositions/useFavicon';
import useNotifications from '~/compositions/useNotifications';
import usePipeline from '~/compositions/usePipeline';
import { useRouteBackOrDefault } from '~/compositions/useRouteBackOrDefault';
import { Repo, RepoPermissions } from '~/lib/api/types';
import PipelineStore from '~/store/pipelines';

export default defineComponent({
  name: 'PipelineWrapper',

  components: {
    Button,
    PipelineStatusIcon,
    Tab,
    Tooltip,
    Scaffold,
  },

  props: {
    repoOwner: {
      type: String,
      required: true,
    },

    repoName: {
      type: String,
      required: true,
    },

    pipelineId: {
      type: String,
      required: true,
    },
  },

  setup(props) {
    const apiClient = useApiClient();
    const route = useRoute();
    const router = useRouter();
    const notifications = useNotifications();
    const favicon = useFavicon();
    const i18n = useI18n();

    const pipelineStore = PipelineStore();
    const pipelineId = toRef(props, 'pipelineId');
    const repoOwner = toRef(props, 'repoOwner');
    const repoName = toRef(props, 'repoName');
    const repo = inject<Ref<Repo>>('repo');
    const repoPermissions = inject<Ref<RepoPermissions>>('repo-permissions');
    if (!repo || !repoPermissions) {
      throw new Error('Unexpected: "repo" & "repoPermissions" should be provided at this place');
    }

    const pipeline = pipelineStore.getPipeline(repoOwner, repoName, pipelineId);
    const { since, duration, created } = usePipeline(pipeline);
    provide('pipeline', pipeline);

    const { message } = usePipeline(pipeline);

    async function loadPipeline(): Promise<void> {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await pipelineStore.loadPipeline(repo.value.owner, repo.value.name, parseInt(pipelineId.value, 10));

      favicon.updateStatus(pipeline.value.status);
    }

    const { doSubmit: cancelPipeline, isLoading: isCancelingPipeline } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      if (!pipeline.value.steps) {
        throw new Error('Unexpected: Pipeline steps not loaded');
      }

      // TODO: is selectedStepId right?
      // const step = findStep(pipeline.value.steps, selectedStepId.value || 2);

      // if (!step) {
      //   throw new Error('Unexpected: Step not found');
      // }

      await apiClient.cancelPipeline(repo.value.owner, repo.value.name, parseInt(pipelineId.value, 10));
      notifications.notify({ title: i18n.t('repo.pipeline.actions.cancel_success'), type: 'success' });
    });

    const { doSubmit: restartPipeline, isLoading: isRestartingPipeline } = useAsyncAction(async () => {
      if (!repo) {
        throw new Error('Unexpected: Repo is undefined');
      }

      await apiClient.restartPipeline(repo.value.owner, repo.value.name, pipelineId.value, { fork: true });
      notifications.notify({ title: i18n.t('repo.pipeline.actions.restart_success'), type: 'success' });
      // TODO: directly send to newest pipeline?
      await router.push({ name: 'repo', params: { repoName: repo.value.name, repoOwner: repo.value.owner } });
    });

    onMounted(loadPipeline);
    watch([repo, pipelineId], loadPipeline);
    onBeforeUnmount(() => {
      favicon.updateStatus('default');
    });

    const activeTab = computed({
      get() {
        if (route.name === 'repo-pipeline-changed-files') {
          return 'changed-files';
        }

        if (route.name === 'repo-pipeline-config') {
          return 'config';
        }

        return 'tasks';
      },
      set(tab: string) {
        if (tab === 'tasks') {
          router.replace({ name: 'repo-pipeline' });
        }

        if (tab === 'changed-files') {
          router.replace({ name: 'repo-pipeline-changed-files' });
        }

        if (tab === 'config') {
          router.replace({ name: 'repo-pipeline-config' });
        }
      },
    });

    return {
      repoPermissions,
      pipeline,
      repo,
      message,
      isCancelingPipeline,
      isRestartingPipeline,
      activeTab,
      since,
      duration,
      cancelPipeline,
      restartPipeline,
      goBack: useRouteBackOrDefault({ name: 'repo' }),
      created,
    };
  },
});
</script>
