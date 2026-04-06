import { proxy } from "valtio"

export const processManagerStore = proxy({
  selectedProcessId: null as string | null,
  terminalProcessId: null as string | null,
  busyActionKey: null as string | null,
  actionError: null as string | null,
})
