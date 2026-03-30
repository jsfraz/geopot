<script setup lang="ts">
import { ref, onMounted } from 'vue';
import { VueSpinnerPulse } from 'vue3-spinners';
import { Configuration } from '@/api/runtime';
import { StatsApi } from '@/api/apis';
import type { ModelsTopEntry } from '@/api/models';

const props = defineProps({
    apiConfiguration: { type: Configuration, required: true },
    spinnerColor: { type: String, default: '#20c20e' },
    limit: { type: Number, default: 8 },
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
    <div class="border border-hacker bg-hackerbg rounded-lg relative w-full h-full flex flex-col overflow-hidden">
        <Transition name="fade">
            <div v-if="!isLoaded"
                class="absolute inset-0 flex justify-center items-center z-10 bg-black/30">
                <VueSpinnerPulse size="10" :color="spinnerColor" />
            </div>
        </Transition>

        <!-- Top Usernames -->
        <div class="flex flex-col flex-1 min-h-0 overflow-hidden p-3 border-b border-hacker/40">
            <p class="text-xs font-medium uppercase text-gray-400 tracking-widest font-mono mb-2 shrink-0">Top Usernames</p>
            <div class="flex-1 overflow-y-auto scrollbar-thin">
                <table class="w-full font-mono text-xs">
                    <thead>
                        <tr class="text-gray-600 border-b border-gray-800">
                            <th class="text-left pb-1 pr-2 font-normal">#</th>
                            <th class="text-left pb-1 pr-2 font-normal">Username</th>
                            <th class="text-right pb-1 pr-2 font-normal">Count</th>
                            <th class="text-right pb-1 font-normal">%</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="(entry, i) in usernames" :key="entry.label"
                            class="border-b border-gray-900/50 hover:bg-hacker/5 transition-colors">
                            <td class="py-0.5 pr-2 text-gray-600">{{ i + 1 }}</td>
                            <td class="py-0.5 pr-2 text-hacker">{{ entry.label }}</td>
                            <td class="py-0.5 pr-2 text-gray-300 text-right">{{ entry.count.toLocaleString() }}</td>
                            <td class="py-0.5 text-gray-500 text-right">{{ entry.percentage.toFixed(1) }}%</td>
                        </tr>
                    </tbody>
                </table>
            </div>
        </div>

        <!-- Top Passwords -->
        <div class="flex flex-col flex-1 min-h-0 overflow-hidden p-3">
            <p class="text-xs font-medium uppercase text-gray-400 tracking-widest font-mono mb-2 shrink-0">Top Passwords</p>
            <div class="flex-1 overflow-y-auto scrollbar-thin">
                <table class="w-full font-mono text-xs">
                    <thead>
                        <tr class="text-gray-600 border-b border-gray-800">
                            <th class="text-left pb-1 pr-2 font-normal">#</th>
                            <th class="text-left pb-1 pr-2 font-normal">Password</th>
                            <th class="text-right pb-1 pr-2 font-normal">Count</th>
                            <th class="text-right pb-1 font-normal">%</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr v-for="(entry, i) in passwords" :key="entry.label"
                            class="border-b border-gray-900/50 hover:bg-red-900/5 transition-colors">
                            <td class="py-0.5 pr-2 text-gray-600">{{ i + 1 }}</td>
                            <td class="py-0.5 pr-2 text-red-400">{{ entry.label }}</td>
                            <td class="py-0.5 pr-2 text-gray-300 text-right">{{ entry.count.toLocaleString() }}</td>
                            <td class="py-0.5 text-gray-500 text-right">{{ entry.percentage.toFixed(1) }}%</td>
                        </tr>
                    </tbody>
                </table>
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
