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
    new StatsApi(props.apiConfiguration).getRecentConnections(props.initialLimit).subscribe({
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
    <div class="border border-hacker bg-hackerbg rounded-lg relative w-full h-full flex flex-col overflow-hidden">
        <!-- Header -->
        <div class="flex items-center gap-2 px-3 pt-3 pb-2 shrink-0 border-b border-hacker/30">
            <span class="relative flex h-2 w-2 shrink-0">
                <span class="animate-ping absolute inline-flex h-full w-full rounded-full bg-hacker opacity-75"></span>
                <span class="relative inline-flex rounded-full h-2 w-2 bg-hacker"></span>
            </span>
            <p class="text-xs font-medium uppercase text-gray-400 tracking-widest font-mono">Live Connection Feed</p>
        </div>

        <!-- Spinner -->
        <Transition name="fade">
            <div v-if="!isLoaded"
                class="absolute inset-0 flex justify-center items-center z-10 bg-black/30">
                <VueSpinnerPulse size="10" :color="spinnerColor" />
            </div>
        </Transition>

        <!-- Table header -->
        <div class="grid grid-cols-[70px_60px_140px_1fr_1fr] gap-x-2 px-3 py-1 shrink-0 text-gray-600 font-mono text-[10px] uppercase tracking-wider border-b border-gray-800">
            <span>Time</span>
            <span>CC</span>
            <span>IP Address</span>
            <span>Username</span>
            <span>Password</span>
        </div>

        <!-- Feed rows -->
        <div ref="feedEl" @scroll="onScroll" class="flex-1 overflow-y-auto min-h-0 scrollbar-thin">
            <div
                v-for="(entry, i) in entries"
                :key="`${entry.id}-${i}`"
                :class="[
                    'grid grid-cols-[70px_60px_140px_1fr_1fr] gap-x-2 px-3 py-0.75 font-mono text-xs border-b border-gray-900/40 transition-colors',
                    entry.isProxy ? 'bg-red-950/20 hover:bg-red-950/30' : (i % 2 === 0 ? 'bg-transparent hover:bg-hacker/5' : 'bg-hacker/2 hover:bg-hacker/5'),
                    i === entries.length - 1 ? 'text-hacker' : ''
                ]"
            >
                <span class="text-gray-500 tabular-nums">{{ formatTime(entry.timestamp) }}</span>
                <span class="text-hacker/80 truncate" :title="entry.countryName">
                    {{ countryFlag(entry.countryCode) }} {{ entry.countryCode || '??' }}
                </span>
                <span :class="['truncate tabular-nums', entry.isProxy ? 'text-red-400' : 'text-gray-300']" :title="entry.ipAddress">
                    {{ entry.ipAddress }}
                </span>
                <span class="text-hacker truncate" :title="entry.user">{{ entry.user || '-' }}</span>
                <span :class="['truncate', entry.isProxy ? 'text-red-300' : 'text-orange-400/80']" :title="entry.password">
                    {{ entry.password || '-' }}
                </span>
            </div>
            <!-- Blinking cursor at end -->
            <div v-if="isLoaded" class="px-3 py-1">
                <span class="font-mono text-xs text-hacker animate-pulse">▮</span>
            </div>
        </div>
    </div>
</template>

<style scoped>
.fade-enter-active, .fade-leave-active { transition: opacity 0.3s ease; }
.fade-enter-from, .fade-leave-to { opacity: 0; }
.scrollbar-thin::-webkit-scrollbar { width: 3px; }
.scrollbar-thin::-webkit-scrollbar-track { background: transparent; }
.scrollbar-thin::-webkit-scrollbar-thumb { background: #20c20e40; border-radius: 2px; }
</style>
