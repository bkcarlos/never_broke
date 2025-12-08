import { get, post, patch } from './http'
import type { Account, CreateAccountPayload, UpdateCreditLimitPayload } from '@/types'

export async function getAccounts(): Promise<Account[]> {
  try {
    return await get<Account[]>('/accounts')
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to fetch accounts')
  }
}

export async function createAccount(
  payload: CreateAccountPayload,
): Promise<Account> {
  try {
    return await post<Account>('/accounts', payload)
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to create account')
  }
}

export async function updateCreditLimits(
  accountId: string,
  payload: UpdateCreditLimitPayload,
): Promise<Account> {
  try {
    return await patch<Account>(`/accounts/${accountId}/credit-limits`, payload)
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to update credit limits')
  }
}
