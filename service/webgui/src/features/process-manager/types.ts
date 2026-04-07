export type ProcessStatus =
  | "respawn"
  | "running"
  | "starting"
  | "stopped"
  | "stopping"
  | "failed"
  | string

export type Process = {
  id: string
  name?: string
  namespace?: string
  exec: string
  args: string[]
  env: string[]
  cwd: string
  uptime?: number
  startCount?: number
  failCount?: number
  status: ProcessStatus
  exitCode?: number
}

export type ProcessAction = "stop" | "restart" | "remove"

export type ApiResponse = {
  result?: {
    processList?: Process[]
  }
  params?: {
    message?: string
  }
}
