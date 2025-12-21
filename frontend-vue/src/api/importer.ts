import { request } from './http'
import type { ImportPreviewResult } from '@/types'

/**
 * 预览导入
 */
export async function previewImport(raw: string): Promise<ImportPreviewResult> {
  try {
    const resp = await request<ImportPreviewResult>('/import/preview', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ raw })
    })
    return resp
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to preview import')
  }
}








