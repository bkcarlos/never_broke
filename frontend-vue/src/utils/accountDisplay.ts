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
    case 'spdb':
      return t('account.institution.spdb', '浦发')
    case 'citic':
      return t('account.institution.citic', '中信')
    case 'ceb':
      return t('account.institution.ceb', '光大')
    case 'hxb':
      return t('account.institution.hxb', '华夏')
    case 'cmbc':
      return t('account.institution.cmbc', '民生')
    case 'cgb':
      return t('account.institution.cgb', '广发')
    case 'pab':
      return t('account.institution.pab', '平安')
    case 'cib':
      return t('account.institution.cib', '兴业')
    case 'psbc':
      return t('account.institution.psbc', '邮储')
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

export function formatAccountDisplayName(account: Pick<Account, 'type' | 'institution' | 'cardLast4'>): string {
  const institutionLabel = getInstitutionLabel(account.institution)
  const typeLabel = getTypeLabel(account.type)

  // 现金类型：只显示类型
  if (account.type === 'cash') {
    return typeLabel
  }

  // 统一格式：机构 + 类型(尾号)
  const parts: string[] = []
  
  // 1. 添加机构/平台（如果有）
  if (institutionLabel) {
    parts.push(institutionLabel)
  }

  // 2. 添加账户类型(尾号)
  if (account.cardLast4 && account.cardLast4.length > 0) {
    parts.push(`${typeLabel}(${account.cardLast4})`)
  } else {
    parts.push(typeLabel)
  }

  return parts.length > 0 ? parts.join(' ') : t('account.display.unnamed', '未配置账户')
}
