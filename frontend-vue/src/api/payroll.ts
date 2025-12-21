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
export async function getPayrollRegionRule(
  code: string,
  year: number = 2025
): Promise<PayrollRegionRule> {
  return get<PayrollRegionRule>(`/payroll/regions/${code}/rule?year=${year}`)
}

/**
 * 计算薪资
 */
export async function calcSalary(
  payload: CalcSalaryRequest
): Promise<CalcSalaryResponse> {
  return post<CalcSalaryResponse>('/payroll/calc/salary', payload)
}
