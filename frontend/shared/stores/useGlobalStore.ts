import { defineStore } from 'pinia'
import { ref } from 'vue'

/**
 * Global store — manages global application state.
 * Holds active contract, business date, and app-wide settings.
 */
export const useGlobalStore = defineStore('global', () => {
  const activeContractId = ref<string | null>(null)
  const businessDate = ref<string | null>(null)
  const isLoading = ref(false)

  function setActiveContract(contractId: string) {
    activeContractId.value = contractId
  }

  function setBusinessDate(date: string) {
    businessDate.value = date
  }

  return {
    activeContractId,
    businessDate,
    isLoading,
    setActiveContract,
    setBusinessDate,
  }
})
