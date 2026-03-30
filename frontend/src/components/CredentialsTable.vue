<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { VueSpinnerPulse } from 'vue3-spinners';
import { Configuration } from '@/api/runtime';
import { StatsApi } from '@/api/apis';
import type { ModelsTopEntry } from '@/api/models';

const props = defineProps({
    apiConfiguration: { type: Configuration, required: true },
    spinnerColor: { type: String, default: '#20c20e' },
    limit: { type: Number, default: 100 },
});

const isLoaded = ref(false);
const usernames = ref<ModelsTopEntry[]>([]);
const passwords = ref<ModelsTopEntry[]>([]);

onMounted(() => {
    const api = new StatsApi(props.apiConfiguration);
    let usersDone = false;
    let passsDone = false;
    const checkDone = () => { if (usersDone && passsDone) isLoaded.value = true; };

    api.getTopUsernames(props.limit).subscribe({
        next: (data) => { usernames.value = data; },
        error: () => { usersDone = true; checkDone(); },
        complete: () => { usersDone = true; checkDone(); },
    });
    api.getTopPasswords(props.limit).subscribe({
        next: (data) => { passwords.value = data; },
        error: () => { passsDone = true; checkDone(); },
        complete: () => { passsDone = true; checkDone(); },
    });
});
</script>

<template>
    <div class="hud-card relative w-full h-full flex flex-col rounded-sm overflow-hidden p-2">
        <Transition name="fade">
            <div v-if="!isLoaded" class="absolute inset-0 flex justify-center items-center z-10 bg-black/50">
                <VueSpinnerPulse size="10" :color="spinnerColor" />
            </div>
        </Transition>

        <!-- Top Usernames -->
        <div class="flex flex-col flex-1 min-h-0 overflow-hidden p-3 border-b border-hacker/20 bg-hacker/5">
            <p class="text-[11px] font-bold uppercase text-hacker/80 tracking-[0.2em] font-mono italic mb-2 shrink-0">
                <span class="mr-1 text-hacker">▶</span> Top Usernames
            </p>
            <div class="flex-1 overflow-y-auto scrollbar-thin">
                <table class="w-full font-mono text-[12px]">
                    <thead>
                        <tr class="text-hacker/60 border-b border-hacker/20 uppercase text-[10px]">
                            <th class="text-left pb-1.5 pr-2 font-normal">#</th>
                            <th class="text-left pb-1.5 pr-2 font-normal">User</th>
                            <th class="text-right pb-1.5 pr-2 font-normal">Total</th>
                            <th class="text-right pb-1.5 font-normal">%</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-hacker/10">
                        <tr v-for="(entry, i) in usernames" :key="entry.label"
                            class="hover:bg-hacker/10 transition-colors">
                            <td class="py-1.5 pr-2 text-gray-500 font-bold">{{ i + 1 }}</td>
                            <td class="py-1.5 pr-2 text-white font-medium">{{ entry.label }}</td>
                            <td class="py-1.5 pr-2 text-hacker/90 text-right tabular-nums font-bold">{{
                                entry.count.toLocaleString() }}</td>
                            <td class="py-1.5 text-hacker/60 text-right tabular-nums">{{ entry.percentage.toFixed(1) }}%
                            </td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <!-- Top Passwords -->
        <div class="flex flex-col flex-1 min-h-0 overflow-hidden p-3 bg-danger/5">
            <p class="text-[11px] font-bold uppercase text-danger/80 tracking-[0.2em] font-mono italic mb-2 shrink-0">
                <span class="mr-1 text-danger">▶</span> Top Passwords
            </p>
            <div class="flex-1 overflow-y-auto scrollbar-thin">
                <table class="w-full font-mono text-[12px]">
                    <thead>
                        <tr class="text-danger/60 border-b border-danger/20 uppercase text-[10px]">
                            <th class="text-left pb-1.5 pr-2 font-normal">#</th>
                            <th class="text-left pb-1.5 pr-2 font-normal">Pass</th>
                            <th class="text-right pb-1.5 pr-2 font-normal">Total</th>
                            <th class="text-right pb-1.5 font-normal">%</th>
                        </tr>
                    </thead>
                    <tbody class="divide-y divide-danger/10">
                        <tr v-for="(entry, i) in passwords" :key="entry.label"
                            class="hover:bg-danger/10 transition-colors">
                            <td class="py-1.5 pr-2 text-gray-500 font-bold">{{ i + 1 }}</td>
                            <td class="py-1.5 pr-2 text-white font-medium">{{ entry.label }}</td>
                            <td class="py-1.5 pr-2 text-danger/90 text-right tabular-nums font-bold">{{
                                entry.count.toLocaleString() }}</td>
                            <td class="py-1.5 text-danger/60 text-right tabular-nums">{{ entry.percentage.toFixed(1) }}%
                            </td>
                        </tr>
                    </tbody>
                </table>
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
