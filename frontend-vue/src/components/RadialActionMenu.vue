<template>
  <teleport to="body">
    <transition name="radial-fade">
      <div
        v-if="visible"
        class="fixed inset-0 z-40 flex items-end justify-center bg-black/30"
        @click.self="emit('close')"
      >
        <!-- 底部中心点容器 -->
        <div class="relative mb-12">
          <!-- 圆弧上的操作按钮 -->
          <div
            v-for="(action, index) in actions"
            :key="action.key"
            class="absolute left-1/2 bottom-0"
            :style="{
              transform: `translate(-50%, 0) translate(${positions[index].x}px, ${positions[index].y}px)`
            }"
          >
            <button
              class="w-20 h-20 rounded-full bg-white shadow-lg flex flex-col items-center justify-center text-xs hover:scale-110 transition-transform"
              @click="onSelect(action)"
            >
              <Icon :name="action.icon" size="28px" class="mb-1" />
              <span class="px-1 text-xs leading-tight font-medium">{{ action.label }}</span>
            </button>
          </div>
        </div>
      </div>
    </transition>
  </teleport>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { RadialAction } from '@/types/actions'
import Icon from './Icon.vue'

interface Props {
  visible: boolean
  actions: RadialAction[]
}

const props = defineProps<Props>()

const emit = defineEmits<{
  (e: 'close'): void
  (e: 'select', action: RadialAction): void
}>()

// 圆弧布局计算
const radius = 100 // px，减小半径让按钮更靠近底部
const startAngle = (210 * Math.PI) / 180 // 弧度，从左下开始
const endAngle = (330 * Math.PI) / 180 // 弧度，到右下结束

const positions = computed(() => {
  const angleStep =
    props.actions.length > 1
      ? (endAngle - startAngle) / (props.actions.length - 1)
      : 0

  return props.actions.map((_, index) => {
    const angle = startAngle + angleStep * index
    const x = radius * Math.cos(angle)
    const y = radius * Math.sin(angle)
    return { x, y }
  })
})

function onSelect(action: RadialAction) {
  emit('select', action)
}
</script>

<style scoped>
.radial-fade-enter-active,
.radial-fade-leave-active {
  transition: opacity 0.3s ease;
}

.radial-fade-enter-from,
.radial-fade-leave-to {
  opacity: 0;
}
</style>

