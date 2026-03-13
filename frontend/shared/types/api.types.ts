/** Standard API response envelope */
export interface ApiResponse<T> {
  data: T
  message?: string
}

/** Standard API error response */
export interface ApiError {
  error: string
  code?: string
  details?: unknown
}

/** Paginated API response */
export interface PaginatedResponse<T> {
  data: T[]
  page: number
  limit: number
  total: number
  total_pages: number
}
