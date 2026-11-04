import { get, post } from './http'
import type {
  PayrollRegion,
  PayrollRegionRule,
  CalcSalaryRequest,
  CalcSalaryResponse
} from '@/types'

/**
 * 获取薪资计算器地区列表
 */
export async function getPayrollRegions(): Promise<PayrollRegion[]> {
  return get<PayrollRegion[]>('/payroll/regions')
}

/**
 * 获取地区规则
 */
export async function getPayrollRegionRule(code: string, date?: string): Promise<PayrollRegionRule> {
  const query = date ? '?date='+encodeURIComponent(date) : ''
  const raw = await get<PayrollRegionRule>(`/payroll/regions/${encodeURIComponent(code)}/rule${query}`)
  // Go decimal fields are serialized as strings. Normalize once at the API boundary.
  const keys = ['ssBaseMin','ssBaseMax','hfBaseMin','hfBaseMax','ssPersonRateSum','ssCompanyRateSum','hfRateMin','hfRateMax','defaultHfRatePerson','defaultHfRateCompany'] as const
  for (const key of keys) {
    raw[key] = Number(raw[key])
    if (!Number.isFinite(raw[key])) throw new Error('缴费基数数据格式无效，请重试')
  }
  return raw
}

/**
 * 计算薪资
 */
export async function calcSalary(
  payload: CalcSalaryRequest
): Promise<CalcSalaryResponse> {
  const response = await post<CalcSalaryResponse>('/payroll/calc/salary', payload)
  for (const key of Object.keys(response.summary) as (keyof CalcSalaryResponse['summary'])[]) response.summary[key] = Number(response.summary[key])
  for (const month of response.months) {
    for (const key of Object.keys(month) as (keyof typeof month)[]) month[key] = Number(month[key])
  }
  response.totals.netYear = Number(response.totals.netYear)
  response.totals.taxYear = Number(response.totals.taxYear)
  return response
}
