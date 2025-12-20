import { ref, onMounted, onBeforeUnmount, computed } from 'vue'

export type Breakpoint = 'xs' | 'sm' | 'md' | 'lg' | 'xl' | '2xl' | '3xl' | '4xl'

/**
 * 响应式断点系统
 * 提供当前屏幕断点和设备类型判断
 */
export function useBreakpoint() {
  const windowWidth = ref(typeof window !== 'undefined' ? window.innerWidth : 0)
  
  // 断点定义（必须与 tailwind.config.js 一致）
  const breakpoints = {
    xs: 320,
    sm: 640,
    md: 768,
    lg: 1024,
    xl: 1280,
    '2xl': 1536,
    '3xl': 1920,
    '4xl': 2560,
  }
  
  const updateWidth = () => {
    windowWidth.value = window.innerWidth
  }
  
  onMounted(() => {
    if (typeof window !== 'undefined') {
      window.addEventListener('resize', updateWidth)
      updateWidth()
    }
  })
  
  onBeforeUnmount(() => {
    if (typeof window !== 'undefined') {
      window.removeEventListener('resize', updateWidth)
    }
  })
  
  /**
   * 当前断点
   */
  const currentBreakpoint = computed<Breakpoint>(() => {
    const width = windowWidth.value
    if (width >= breakpoints['4xl']) return '4xl'
    if (width >= breakpoints['3xl']) return '3xl'
    if (width >= breakpoints['2xl']) return '2xl'
    if (width >= breakpoints.xl) return 'xl'
    if (width >= breakpoints.lg) return 'lg'
    if (width >= breakpoints.md) return 'md'
    if (width >= breakpoints.sm) return 'sm'
    return 'xs'
  })
  
  /**
   * 设备类型判断
   */
  const isMobile = computed(() => windowWidth.value < breakpoints.md)
  const isTablet = computed(() => 
    windowWidth.value >= breakpoints.md && windowWidth.value < breakpoints.lg
  )
  const isDesktop = computed(() => windowWidth.value >= breakpoints.lg)
  const is2K = computed(() => windowWidth.value >= breakpoints['3xl'])
  const is4K = computed(() => windowWidth.value >= breakpoints['4xl'])
  
  /**
   * 精细化判断
   */
  const isSmallMobile = computed(() => windowWidth.value < breakpoints.sm) // < 640px
  const isLargeMobile = computed(() => 
    windowWidth.value >= breakpoints.sm && windowWidth.value < breakpoints.md
  ) // 640px - 767px
  const isTabletPortrait = computed(() => 
    windowWidth.value >= breakpoints.md && windowWidth.value < breakpoints.lg
  ) // 768px - 1023px
  const isTabletLandscape = computed(() => 
    windowWidth.value >= breakpoints.lg && windowWidth.value < breakpoints.xl
  ) // 1024px - 1279px
  const isLaptop = computed(() => 
    windowWidth.value >= breakpoints.xl && windowWidth.value < breakpoints['2xl']
  ) // 1280px - 1535px
  const isLargeScreen = computed(() => windowWidth.value >= breakpoints['2xl']) // >= 1536px
  
  /**
   * 工具函数：判断是否大于等于某个断点
   */
  const isGreaterOrEqual = (breakpoint: Breakpoint) => {
    return computed(() => windowWidth.value >= breakpoints[breakpoint])
  }
  
  /**
   * 工具函数：判断是否小于某个断点
   */
  const isLessThan = (breakpoint: Breakpoint) => {
    return computed(() => windowWidth.value < breakpoints[breakpoint])
  }
  
  /**
   * 安全区域检测
   */
  const hasSafeArea = computed(() => {
    if (typeof window === 'undefined') return false
    // 检测是否支持安全区域
    const style = getComputedStyle(document.documentElement)
    const top = style.getPropertyValue('env(safe-area-inset-top)')
    return top && parseInt(top) > 0
  })
  
  return {
    // 基础数据
    windowWidth,
    currentBreakpoint,
    
    // 设备类型
    isMobile,
    isTablet,
    isDesktop,
    is2K,
    is4K,
    
    // 精细化判断
    isSmallMobile,
    isLargeMobile,
    isTabletPortrait,
    isTabletLandscape,
    isLaptop,
    isLargeScreen,
    
    // 工具函数
    isGreaterOrEqual,
    isLessThan,
    
    // 安全区域
    hasSafeArea,
  }
}
