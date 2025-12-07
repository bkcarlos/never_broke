import { get, patch } from './http'
import type { UserSettings } from '@/types'

/**
 * 获取用户设置
 */
export function getUserSettings(): Promise<UserSettings> {
  return get<UserSettings>('/user/settings')
}

/**
 * 更新用户设置
 * @param settings 要更新的设置（部分更新）
 */
export function updateUserSettings(settings: Partial<UserSettings>): Promise<UserSettings> {
  return patch<UserSettings>('/user/settings', settings)
}

