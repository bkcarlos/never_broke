import { get } from './http'
import type { Account } from '@/types'

/**
 * 获取账户列表
 */
export function getAccounts(): Promise<Account[]> {
  return get<Account[]>('/accounts')
}
