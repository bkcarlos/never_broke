export type RadialActionKey =
  | 'record_expense'
  | 'record_income'
  | 'record_transfer'
  | 'go_detail'
  | 'go_reports'
  | 'go_profile'

export interface RadialAction {
  key: RadialActionKey
  label: string
  icon: string
  route: string
}

