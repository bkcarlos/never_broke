import { request } from './http'
import type { CitySocialRate, CitySocialRateDetail } from '@/types'

// 获取所有城市的社保比例列表
export async function listCitySocialRates(): Promise<CitySocialRate[]> {
  try {
    return await request<CitySocialRate[]>('/city-social-rates', {
      method: 'GET',
    })
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to fetch city social rates')
  }
}

// 获取指定城市的社保比例详情
export async function getCitySocialRateDetail(cityCode: string): Promise<CitySocialRateDetail> {
  try {
    return await request<CitySocialRateDetail>(`/city-social-rates/${encodeURIComponent(cityCode)}`, {
      method: 'GET',
    })
  } catch (error) {
    throw new Error(error instanceof Error ? error.message : 'Failed to fetch city social rate detail')
  }
}
