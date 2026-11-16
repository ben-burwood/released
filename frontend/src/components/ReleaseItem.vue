<template>
    <!-- Date column (8rem) + gap (2rem); the timeline runs down the middle of the gap -->
    <li class="relative grid grid-cols-[minmax(0,1fr)] gap-1 pb-6 sm:grid-cols-[8rem_1fr] sm:gap-8">
        <!-- TIMELINE -->
        <span class="absolute top-0 bottom-0 left-36 hidden w-0.5 -translate-x-1/2 bg-mist-400 sm:block"></span>
        <span class="absolute top-4 left-36 hidden size-4 -translate-x-1/2 rounded-full border-3 border-mist-200 ring-2 sm:block" :class="dotClass"></span>

        <!-- DATE -->
        <div class="leading-snug sm:pt-3 sm:text-right">
            <span class="block text-sm font-semibold" :class="release.draft ? 'text-gray-600' : 'text-navy-900'">{{ dateBig }}</span>
            <span v-if="dateSmall" class="block font-mono text-xs text-gray-400">{{ dateSmall }}</span>
        </div>

        <!-- CARD -->
        <article
            class="min-w-0 rounded-xl border px-5 pt-3.5 pb-2"
            :class="cardClass"
        >
            <header class="flex flex-wrap items-start justify-between gap-3">
                <div class="flex flex-col gap-0.5">
                    <span class="text-lg font-semibold tracking-tight" :class="release.draft ? 'text-gray-600' : 'text-navy-900'">{{ release.name }}</span>
                    <span v-if="release.author" class="text-xs text-gray-600">by {{ release.author }}</span>
                </div>
                <div class="flex flex-col items-end gap-1">
                    <div class="flex flex-wrap items-center justify-end gap-1.5">
                        <span v-if="latest" :class="BADGE" class="bg-brand-500 text-brand-900">Latest</span>
                        <span v-if="release.draft" :class="BADGE" class="bg-gray-900 text-white">Draft</span>
                        <span class="rounded-md border border-mist-200 bg-mist-50 px-2 py-0.5 font-mono text-xs text-navy-700">{{ release.tag }}</span>
                        <span v-if="release.prerelease" :class="BADGE" class="bg-brand-50 text-brand-700">Pre-release</span>
                    </div>
                    <a
                        :href="release.url"
                        target="_blank"
                        rel="noopener"
                        title="View on GitHub"
                        aria-label="View on GitHub"
                        class="grid size-7 place-items-center rounded-md text-gray-400 hover:bg-mist-50 hover:text-navy-900"
                    >
                        <IconGithub class="size-4" aria-hidden="true" />
                    </a>
                </div>
            </header>
            <div v-if="release.bodyHtml" class="release-body mt-2.5" v-html="release.bodyHtml"></div>
        </article>
    </li>
</template>

<script setup lang="ts">
import type { Release } from "@/types/release";
import { formatDate, formatRelative, formatTime } from "@/utils/format";
import { computed } from "vue";
import IconGithub from "~icons/simple-icons/github";

const props = defineProps<{ release: Release; latest: boolean; now: Date }>();

const BADGE = "rounded-full px-2.5 py-0.5 font-mono text-[0.68rem] font-semibold uppercase tracking-wider";

const dotClass = computed(() => {
    if (props.latest) return "bg-brand-500 ring-brand-200";
    if (props.release.draft) return "bg-gray-400 ring-mist-200";
    return "bg-navy-900 ring-mist-400";
});

const cardClass = computed(() => {
    if (props.release.draft) return "border-dashed border-gray-300 bg-gray-100 opacity-85";
    if (props.latest) return "border-2 border-brand-500/50 bg-white shadow-md shadow-brand-500/10";
    return "border-gray-200 bg-white shadow-sm shadow-navy-900/10";
});

const published = computed(() => (props.release.publishedAt ? new Date(props.release.publishedAt) : null));
const dateBig = computed(() => (published.value ? formatDate(published.value) : "—"));
const dateSmall = computed(() => (published.value ? `${formatTime(published.value)} · ${formatRelative(published.value, props.now)}` : ""));
</script>
