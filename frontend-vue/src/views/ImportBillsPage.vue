<template>
  <div class="page-container">
    <div class="page-content px-4">
      <!-- 顶部栏 -->
      <div class="flex items-center gap-4 mb-6">
        <button
          @click="handleBack"
          class="p-2 hover:bg-gray-100 rounded-lg transition-colors"
        >
          <Icon name="arrow-left" size="24px" class="text-gray-700" />
        </button>
        <h1 class="text-2xl font-bold text-gray-900">{{ t('import.title') }}</h1>
      </div>

      <!-- 说明卡片 -->
      <div class="bg-blue-50 rounded-lg shadow p-4 mb-4">
        <p class="text-sm text-gray-700">
          {{ t('import.description') }}
        </p>
      </div>

      <!-- 文本域 -->
      <div class="bg-white rounded-lg shadow p-4 mb-4">
        <textarea
          v-model="rawText"
          :placeholder="t('import.placeholder')"
          class="w-full h-48 px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
        />
      </div>

      <!-- 解析预览按钮 -->
      <div class="mb-4">
        <button
          @click="handlePreview"
          :disabled="loading"
          class="w-full py-3 bg-blue-500 text-white rounded-lg font-semibold hover:bg-blue-600 disabled:bg-gray-300 disabled:cursor-not-allowed transition-colors"
        >
          {{ loading ? t('import.parsing') : t('import.parse') }}
        </button>
      </div>

      <!-- 错误提示 -->
      <div v-if="error" class="mb-4 p-4 bg-red-50 text-red-600 rounded-lg">
        {{ error }}
      </div>

      <!-- 预览区域 -->
      <div v-if="previewResult" class="bg-white rounded-lg shadow p-4">
        <div class="mb-4">
          <p class="text-lg font-semibold text-gray-900">
            {{ t('import.result', { count: previewResult.count }) }}
          </p>
        </div>

        <div class="space-y-3">
          <div
            v-for="(item, index) in previewResult.items"
            :key="index"
            class="border border-gray-200 rounded-lg p-4"
          >
            <div class="flex justify-between items-start mb-2">
              <div>
                <p class="font-medium text-gray-900">{{ item.remark }}</p>
                <p class="text-sm text-gray-500">{{ item.date }}</p>
              </div>
              <p class="text-lg font-bold text-red-600">¥{{ item.amount.toFixed(2) }}</p>
            </div>
            <div class="mt-2">
              <span class="inline-block px-2 py-1 bg-gray-100 text-gray-600 text-xs rounded">
                {{ item.categoryGuess }}
              </span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { previewImport } from '@/api/importer'
import type { ImportPreviewResult } from '@/types'
import Icon from '@/components/Icon.vue'
import { t } from '@/i18n'

const router = useRouter()

const rawText = ref<string>('')
const loading = ref(false)
const error = ref<string | null>(null)
const previewResult = ref<ImportPreviewResult | null>(null)

function handleBack() {
  router.back()
}

async function handlePreview() {
  if (!rawText.value.trim()) {
    error.value = t('import.error.empty')
    return
  }

  loading.value = true
  error.value = null
  previewResult.value = null

  try {
    const result = await previewImport(rawText.value)
    previewResult.value = result
  } catch (err) {
    error.value = err instanceof Error ? err.message : t('import.error.parse')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
/* 页面样式已移至全局样式文件 */
</style>




