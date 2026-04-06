import { useMemo } from "react"
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query"
import { useSnapshot } from "valtio"
import { createProcessManagerApi } from "./api"
import { processManagerStore } from "./store"
import type { ProcessAction } from "./types"
import { readTokenFromHash } from "./utils"

export function useProcessManager() {
  const queryClient = useQueryClient()
  const ui = useSnapshot(processManagerStore)

  const token = useMemo(() => readTokenFromHash(window.location.hash), [])
  const api = useMemo(() => createProcessManagerApi(token), [token])

  const processesQuery = useQuery({
    queryKey: ["processes", token],
    queryFn: () => api.listProcesses(),
    refetchInterval: 3000,
  })

  const actionMutation = useMutation({
    mutationFn: async ({ action, processId }: { action: ProcessAction; processId: string }) => {
      await api.runAction(action, processId)
    },
    onMutate: ({ action, processId }) => {
      processManagerStore.busyActionKey = `${action}:${processId}`
      processManagerStore.actionError = null
    },
    onError: (err) => {
      processManagerStore.actionError = err instanceof Error ? err.message : "Failed to run action"
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ["processes", token] })
    },
    onSettled: () => {
      processManagerStore.busyActionKey = null
    },
  })

  const fetchProcesses = () => {
    void processesQuery.refetch()
  }

  const runAction = (action: ProcessAction, processId: string) => {
    actionMutation.mutate({ action, processId })
  }

  const processes = processesQuery.data ?? []

  const queryError = processesQuery.error instanceof Error ? processesQuery.error.message : null
  const error = ui.actionError ?? queryError

  const selectedProcess = useMemo(
    () => processes.find((p) => p.id === ui.selectedProcessId) ?? null,
    [processes, ui.selectedProcessId],
  )

  const terminalProcess = useMemo(
    () => processes.find((p) => p.id === ui.terminalProcessId) ?? null,
    [processes, ui.terminalProcessId],
  )

  return {
    token,
    processes,
    isLoading: processesQuery.isLoading,
    error,
    busyActionKey: ui.busyActionKey,
    selectedProcess,
    terminalProcess,
    fetchProcesses,
    runAction,
    setSelectedProcessId: (processId: string | null) => {
      processManagerStore.selectedProcessId = processId
    },
    setTerminalProcessId: (processId: string | null) => {
      processManagerStore.terminalProcessId = processId
    },
  }
}
