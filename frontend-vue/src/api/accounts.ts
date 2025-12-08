import { get, post } from './http'
import type { Account, CreateAccountPayload } from '@/types'

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

/**
 * 创建账户
 */
export async function createAccount(payload: CreateAccountPayload): Promise<Account> {
  try {
    return await post<Account>('/accounts', payload)
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to create account')
  }
}
