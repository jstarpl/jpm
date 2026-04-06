import type { ApiResponse, Process, ProcessAction } from "./types"

function buildHeaders(token: string, initHeaders?: HeadersInit): Headers {
  const headers = new Headers(initHeaders)
  if (token) {
    headers.set("Authorization", `Bearer ${token}`)
  }
  return headers
}

export function createProcessManagerApi(token: string) {
  const apiRequest = async (path: string, init?: RequestInit) => {
    const response = await fetch(`/api${path}`, {
      ...init,
      headers: buildHeaders(token, init?.headers),
    })

    if (!response.ok) {
      let details = `${response.status} ${response.statusText}`
      try {
        const data = (await response.json()) as ApiResponse
        details = data.params?.message || details
      } catch {
        // Ignore JSON parse errors for non-JSON responses.
      }
      throw new Error(details)
    }

    if (response.status === 204) {
      return null
    }

    const contentType = response.headers.get("content-type") || ""
    if (!contentType.includes("application/json")) {
      return null
    }

    return (await response.json()) as ApiResponse
  }

  return {
    async listProcesses(): Promise<Process[]> {
      const data = await apiRequest("/processes", { method: "GET" })
      return data?.result?.processList ?? []
    },
    async runAction(action: ProcessAction, processId: string): Promise<void> {
      if (action === "remove") {
        await apiRequest(`/processes/${processId}`, { method: "DELETE" })
        return
      }

      await apiRequest(`/processes/${processId}/${action}`, { method: "POST" })
    },
    async sendStdin(processId: string, value: string): Promise<void> {
      await fetch(`/api/processes/${processId}/stdin`, {
        method: "POST",
        headers: buildHeaders(token),
        body: value,
      })
    },
  }
}
