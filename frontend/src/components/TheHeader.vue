<script setup lang="ts">
defineProps<{
  isAuthenticated: boolean
  lastUpdated: string
}>()
</script>

<template>
  <header class="flex items-center justify-between p-4 bg-slate-950 border-b border-slate-800 shadow-2xl">
    <h1 class="text-5xl font-black text-white tracking-tighter italic shrink-0">
      FITBITER
    </h1>

    <div class="flex items-center gap-6 shrink-0">

      <div class="flex items-center gap-3 px-4 py-2 bg-slate-900 rounded-lg border border-slate-700">
        <div class="relative flex h-2.5 w-2.5">
          <span v-if="isAuthenticated"
            class="animate-ping absolute inline-flex h-full w-full rounded-full bg-green-400 opacity-75"></span>
          <span :class="isAuthenticated ? 'bg-green-500' : 'bg-red-500'"
            class="relative inline-flex rounded-full h-2.5 w-2.5"></span>
        </div>
        <span class="text-sm font-bold text-slate-200 tracking-wider uppercase">
          {{ isAuthenticated ? 'Connected' : 'Disconnected' }}
        </span>
      </div>

      <div class="flex flex-col items-end border-l border-slate-800 pl-6">
        <span class="text-sm text-slate-600 font-bold uppercase tracking-widest leading-none mb-1">
          Last Synced
        </span>
        <span class="text-sm text-slate-400 font-mono leading-none">
          {{ lastUpdated || '---' }}
        </span>
      </div>

    </div>
  </header>
</template>
