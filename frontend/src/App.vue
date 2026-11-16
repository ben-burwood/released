<template>
    <div class="min-h-screen w-full">
        <main class="mx-auto max-w-5xl px-4 py-10 sm:px-6">
            <h1 class="mb-8 text-3xl font-bold tracking-tight text-navy-950">
                <span v-if="repos.length > 1" class="relative inline-block max-w-full align-baseline">
                    <select
                        v-model="repo"
                        aria-label="Repository"
                        class="max-w-full cursor-pointer appearance-none truncate rounded-lg border border-transparent bg-transparent py-0.5 pr-9 pl-1 font-[inherit] hover:border-mist-400 focus:border-brand-500 focus:outline-none"
                        @change="selectRepo"
                    >
                        <option v-for="r in repos" :key="r" :value="r" class="text-base font-normal">{{ repoLabel(r) }}</option>
                    </select>
                    <IconChevronDown class="pointer-events-none absolute top-1/2 right-2 size-5 -translate-y-1/2 text-gray-500" aria-hidden="true" />
                </span>
                <template v-else>{{ repoName }}</template>
            </h1>

            <div v-if="loading" class="flex items-center gap-3 text-sm text-gray-600">
                <span class="size-4 animate-spin rounded-full border-2 border-mist-400 border-t-brand-500"></span>
                Loading releases…
            </div>
            <div v-else-if="error" class="rounded-lg border border-red-200 bg-red-50 px-4 py-3 text-sm text-red-800">Failed to load releases: {{ error }}</div>
            <div v-else-if="!releases.length" class="rounded-lg bg-mist-300/60 px-4 py-3 text-sm text-navy-900">No releases found.</div>

            <template v-else>
                <!-- TIMELINE -->
                <ol class="mt-2 mb-10">
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
import IconChevronDown from "~icons/lucide/chevron-down";

const PAGE_SIZE = 25;

const REPO_PARAM = "repo";

const repos = ref<string[]>([]);
const repo = ref("");
const releases = ref<Release[]>([]);
const loading = ref(true);
const error = ref("");
const shown = ref(PAGE_SIZE);

const repoName = computed(() => repo.value.split("/").pop() ?? "");
// Labels a repo by its name, or by owner/name when another configured repo shares the name
function repoLabel(r: string) {
    const name = r.split("/").pop() ?? r;
    return repos.value.filter((o) => o.split("/").pop()?.toLowerCase() === name.toLowerCase()).length > 1 ? r : name;
}

const shownCount = computed(() => Math.min(shown.value, releases.value.length));

// Latest is the newest full release, falling back to the first listed
const latest = computed(() => releases.value.find((r) => !r.prerelease && !r.draft) ?? releases.value[0]);

// Ticks relative times ("5m ago")
const now = ref(new Date());
const ticker = setInterval(() => (now.value = new Date()), 30_000);

async function getJSON<T>(url: string): Promise<T> {
    const resp = await fetch(url);
    if (!resp.ok) throw new Error((await resp.text()).trim() || resp.statusText);
    return resp.json();
}

// Counts loads, so a slow response for a previously selected repo is dropped
let loadID = 0;

// Loads the releases of ?repo= (owner/name), or of the server's default repo when unset
async function loadReleases() {
    const id = ++loadID;
    const requested = new URLSearchParams(location.search).get(REPO_PARAM) ?? "";
    loading.value = true;
    error.value = "";
    releases.value = [];
    shown.value = PAGE_SIZE;
    // Show the requested repo while loading, matched case-insensitively against the configured repos
    repo.value = repos.value.find((r) => r.toLowerCase() === requested.toLowerCase()) ?? requested;
    try {
        const query = requested ? `?${REPO_PARAM}=${encodeURIComponent(requested)}` : "";
        const data = await getJSON<Releases>(`${SERVER_URL}/releases${query}`);
        if (id !== loadID) return;
        repo.value = data.repo;
        releases.value = data.releases;
    } catch (e) {
        if (id !== loadID) return;
        error.value = e instanceof Error ? e.message : String(e);
    }
    loading.value = false;
}

function selectRepo() {
    const url = new URL(location.href);
    url.searchParams.set(REPO_PARAM, repo.value);
    history.pushState(null, "", url);
    loadReleases();
}

onMounted(async () => {
    window.addEventListener("popstate", loadReleases);
    try {
        repos.value = await getJSON<string[]>(`${SERVER_URL}/repos`);
    } catch {
        // The selector is optional; releases still load for ?repo= or the default
    }
    await loadReleases();
});

onUnmounted(() => {
    clearInterval(ticker);
    window.removeEventListener("popstate", loadReleases);
});
</script>
