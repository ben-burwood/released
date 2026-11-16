<template>
    <div class="min-h-screen w-full">
        <main class="mx-auto max-w-5xl px-4 py-10 sm:px-6">
            <h1 class="mb-8 text-3xl font-bold tracking-tight text-navy-950">Releases : {{ repoName }}</h1>

            <div v-if="loading" class="flex items-center gap-3 text-sm text-gray-600">
                <span class="size-4 animate-spin rounded-full border-2 border-mist-400 border-t-brand-500"></span>
                Loading releases…
            </div>
            <div v-else-if="error" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">Failed to load releases: {{ error }}</div>
            <div v-else-if="!releases.length" class="rounded-lg bg-mist-300/60 px-4 py-3 text-sm text-navy-900">No releases found.</div>

            <template v-else>
                <!-- TIMELINE -->
                <ol
                    class="relative mt-2 mb-10 before:absolute before:top-2.5 before:bottom-2.5 before:left-[9.05rem] before:hidden before:w-0.5 before:bg-linear-to-b before:from-navy-900 before:to-mist-400 sm:before:block"
                >
                    <ReleaseItem v-for="r in releases.slice(0, shownCount)" :key="r.url" :release="r" :latest="r === latest" :now="now" />
                </ol>

                <p class="text-sm text-gray-500">Showing {{ shownCount }} of {{ releases.length }} releases</p>
                <button
                    v-if="shownCount < releases.length"
                    class="mt-3 cursor-pointer rounded-lg border border-mist-400 bg-white px-4 py-2 text-sm font-medium text-navy-900 hover:border-brand-500 hover:text-brand-700"
                    @click="shown += PAGE_SIZE"
                >
                    Show {{ Math.min(PAGE_SIZE, releases.length - shownCount) }} more
                </button>
            </template>
        </main>
    </div>
</template>

<script setup lang="ts">
import ReleaseItem from "@/components/ReleaseItem.vue";
import { SERVER_URL } from "@/main";
import type { Release, Releases } from "@/types/release";
import { computed, onMounted, onUnmounted, ref } from "vue";

const PAGE_SIZE = 25;

const repo = ref("");
const releases = ref<Release[]>([]);
const loading = ref(true);
const error = ref("");
const shown = ref(PAGE_SIZE);

const repoName = computed(() => repo.value.split("/").pop() ?? "");
const shownCount = computed(() => Math.min(shown.value, releases.value.length));

// Latest is the newest full release, falling back to the first listed
const latest = computed(() => releases.value.find((r) => !r.prerelease && !r.draft) ?? releases.value[0]);

// Ticks relative times ("5m ago")
const now = ref(new Date());
const ticker = setInterval(() => (now.value = new Date()), 30_000);

onMounted(async () => {
    try {
        const resp = await fetch(`${SERVER_URL}/releases`);
        if (!resp.ok) throw new Error((await resp.text()).trim() || resp.statusText);
        const data: Releases = await resp.json();
        repo.value = data.repo;
        releases.value = data.releases;
    } catch (e) {
        error.value = e instanceof Error ? e.message : String(e);
    } finally {
        loading.value = false;
    }
});

onUnmounted(() => clearInterval(ticker));
</script>
