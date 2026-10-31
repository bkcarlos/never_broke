export function calendarTitle(title: string, kind: string): string {
  if (kind !== 'schedule') return title
  // Generated repayment descriptions include a status and amount already shown in the row.
  return title.replace(/^(?:预计|预计还款|账单到期|还款)\s*[｜|:：]\s*/, '')
    .replace(/\s*(?:¥|￥|US\$|\$|CNY|USD)\s*[\d,]+(?:\.\d+)?\s*$/, '').trim() || '还款安排'
}

export function calendarKind(kind: string, status?: string): string {
  return kind === 'schedule' ? (status === 'billed' ? '还款' : '预计还款') : kind === 'refund' ? '退款' : '消费'
}
