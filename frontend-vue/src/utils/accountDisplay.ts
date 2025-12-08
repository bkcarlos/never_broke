import type { Account, AccountInstitution, AccountType } from '@/types'
import { t } from '@/i18n'

function getInstitutionLabel(inst?: AccountInstitution): string | undefined {
  if (!inst) return undefined

  switch (inst) {
    case 'icbc':
      return t('account.institution.icbc', '工行')
    case 'ccb':
      return t('account.institution.ccb', '建行')
    case 'abc':
      return t('account.institution.abc', '农行')
    case 'boc':
      return t('account.institution.boc', '中行')
    case 'bocom':
      return t('account.institution.bocom', '交行')
    case 'cmb':
      return t('account.institution.cmb', '招行')
    case 'wechat':
      return t('account.institution.wechat', '微信')
    case 'alipay':
      return t('account.institution.alipay', '支付宝')
    case 'other':
      return t('account.institution.other', '其他')
  }
}

function getTypeLabel(type: AccountType): string {
  switch (type) {
    case 'cash':
      return t('account.type.cash', '现金')
    case 'bank':
      return t('account.type.bank', '储蓄卡')
    case 'wallet':
      return t('account.type.wallet', '钱包')
    case 'credit':
      return t('account.type.credit', '信用卡')
  }
}

export function formatAccountDisplayName(account: Account): string {
  const institutionLabel = getInstitutionLabel(account.institution)
  const typeLabel = getTypeLabel(account.type)
  const parts: string[] = []

  // 1. 机构（如果有）
  if (institutionLabel) {
    parts.push(institutionLabel)
  }

  // 2. 账户类型（必定有）
  parts.push(typeLabel)

  // 3. 尾号
  if (account.cardLast4 && account.cardLast4.length > 0) {
    parts.push(
      t('account.display.last4', '尾号{last4}').replace('{last4}', account.cardLast4)
    )
  }

  if (parts.length === 0) {
    return t('account.display.unnamed', '未配置账户')
  }

  return parts.join(' ')
}
