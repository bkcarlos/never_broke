import { get } from './http'
import type { Account, ApiResponse } from '@/types'

/**
 * 获取账户列表
 */
export async function getAccounts(): Promise<Account[]> {
  try {
    return await get<Account[]>('/accounts')
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to fetch accounts')
  }
}
