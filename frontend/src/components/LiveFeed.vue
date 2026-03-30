<script setup lang="ts">
import { ref, onMounted, nextTick } from 'vue';
import { VueSpinnerPulse } from 'vue3-spinners';
import { Configuration } from '@/api/runtime';
import { StatsApi } from '@/api/apis';
import type { ModelsConnection } from '@/api/models';

const props = defineProps({
    apiConfiguration: { type: Configuration, required: true },
    spinnerColor: { type: String, default: '#20c20e' },
    initialLimit: { type: Number, default: 100 },
    maxEntries: { type: Number, default: 200 },
});

defineExpose({ addEntry });

const isLoaded = ref(false);
const feedEl = ref<HTMLElement | null>(null);
const entries = ref<ModelsConnection[]>([]);
let autoScroll = true;

// Country code → flag emoji
function countryFlag(code: string): string {
    if (!code || code.length !== 2) return '🌐';
    return code.toUpperCase().replace(/./g, c =>
        String.fromCodePoint(c.charCodeAt(0) + 127397)
    );
}

function formatTime(ts: string): string {
    try {
        const d = new Date(ts);
        return d.toLocaleTimeString('cs-CZ', { hour: '2-digit', minute: '2-digit', second: '2-digit' });
    } catch {
        return '--:--:--';
    }
}

function scrollToBottom() {
    if (autoScroll && feedEl.value) {
        feedEl.value.scrollTop = feedEl.value.scrollHeight;
    }
}

function onScroll() {
    if (!feedEl.value) return;
    const { scrollTop, scrollHeight, clientHeight } = feedEl.value;
    // Re-enable auto-scroll when user scrolls back to bottom (within 40px)
    autoScroll = scrollHeight - scrollTop - clientHeight < 40;
}

// Add a new real-time entry (called from App.vue via WS)
function addEntry(connection: ModelsConnection) {
    entries.value.push(connection);
    // Keep max entries
    if (entries.value.length > props.maxEntries) {
        entries.value.splice(0, entries.value.length - props.maxEntries);
    }
    nextTick(scrollToBottom);
}

onMounted(() => {
    new StatsApi(props.apiConfiguration).getRecentConnections({ limit: props.initialLimit }).subscribe({
        next: (data: ModelsConnection[]) => {
            // getRecentConnections returns newest-first, reverse for chronological order
            entries.value = [...data].reverse();
        },
        error: (err) => {
            console.error('LiveFeed error:', err);
            isLoaded.value = true;
        },
        complete: () => {
            isLoaded.value = true;
            nextTick(scrollToBottom);
        },
    });
});
</script>

<template>
    <div class="hud-card relative w-full h-full flex flex-col rounded-sm overflow-hidden p-2">
        <!-- Header -->
        <div class="flex items-center gap-2 px-2 pt-2 pb-2 shrink-0 border-b border-hacker/10 mb-1">
            <span class="relative flex h-2 w-2 shrink-0">
                <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-hacker opacity-75"></span>
                <span class="relative inline-flex rounded-full h-2 w-2 bg-hacker shadow-[0_0_5px_#20c20e]"></span>
            </span>
            <p class="text-[10px] font-bold uppercase text-hacker/60 tracking-[0.2em] font-mono italic">
                <span class="mr-1 text-hacker">▶</span> Real-time Feed
            </p>
        </div>

        <!-- Spinner -->
        <Transition name="fade">
            <div v-if="!isLoaded" class="absolute inset-0 flex justify-center items-center z-10 bg-black/50">
                <VueSpinnerPulse size="10" :color="spinnerColor" />
            </div>
        </Transition>

        <!-- Table header -->
        <div
            class="grid grid-cols-[75px_50px_140px_1fr_1fr] gap-x-2 px-3 py-1.5 shrink-0 text-hacker/60 font-mono text-[10px] uppercase tracking-widest border-b border-hacker/20 bg-hacker/5">
            <span>Time</span>
            <span>Origin</span>
            <span>IP Address</span>
            <span>User</span>
            <span>Pass</span>
        </div>

        <!-- Feed rows -->
        <div ref="feedEl" @scroll="onScroll" class="flex-1 overflow-y-auto min-h-0 scrollbar-thin px-1">
            <div v-for="(entry, i) in entries" :key="`${entry.id}-${i}`" :class="[
                'grid grid-cols-[75px_50px_140px_1fr_1fr] gap-x-2 px-2 py-1.5 font-mono text-[13px] transition-all duration-200 border-l-2',
                entry.isProxy ? 'border-danger/60 bg-danger/10 hover:bg-danger/20' : (i % 2 === 0 ? 'border-transparent hover:bg-hacker/10' : 'border-transparent bg-hacker/5 hover:bg-hacker/10'),
                i === entries.length - 1 ? 'glow-text-green text-hacker font-bold' : ''
            ]">
                <span class="text-gray-500 tabular-nums">{{ formatTime(entry.timestamp) }}</span>
                <span class="text-hacker/90 truncate font-bold" :title="entry.countryName">
                    {{ countryFlag(entry.countryCode) }} {{ entry.countryCode || '??' }}
                </span>
                <span
                    :class="['truncate tabular-nums font-bold', entry.isProxy ? 'text-danger glow-text-red' : 'text-gray-200']"
                    :title="entry.ipAddress">
                    {{ entry.ipAddress }}
                </span>
                <span class="text-hacker truncate" :title="entry.user">{{ entry.user || '---' }}</span>
                <span :class="['truncate font-medium', entry.isProxy ? 'text-warning' : 'text-warning/80']"
                    :title="entry.password">
                    {{ entry.password || '---' }}
                </span>
            </div>
            <!-- Blinking cursor at end -->
            <div v-if="isLoaded" class="px-2 py-2">
                <span class="font-mono text-sm text-hacker animate-pulse shadow-[0_0_10px_rgba(57,255,20,0.6)]">▮</span>
            </div>
        </div>
    </div>
</template>

<style scoped>
.fade-enter-active,
.fade-leave-active {
    transition: opacity 0.3s ease;
}

.fade-enter-from,
.fade-leave-to {
    opacity: 0;
}

.scrollbar-thin::-webkit-scrollbar {
    width: 3px;
}

.scrollbar-thin::-webkit-scrollbar-track {
    background: transparent;
}

.scrollbar-thin::-webkit-scrollbar-thumb {
    background: #20c20e40;
    border-radius: 2px;
}
</style>
