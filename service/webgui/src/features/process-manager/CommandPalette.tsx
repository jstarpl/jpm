import { useEffect } from "react"
import { SquareTerminal, RotateCcw, Square, Trash2 } from "lucide-react"
import {
  CommandDialog,
  CommandEmpty,
  CommandGroup,
  CommandInput,
  CommandItem,
  CommandList,
  CommandSeparator,
} from "@/components/ui/command"
import type { Process, ProcessAction } from "./types"

type CommandPaletteProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  processes: Process[]
  busyActionKey: string | null
  onAction: (action: ProcessAction, processId: string) => void
  onOpenTerminal: (processId: string) => void
}

export function CommandPalette({
  open,
  onOpenChange,
  processes,
  busyActionKey,
  onAction,
  onOpenTerminal,
}: CommandPaletteProps) {
  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if (e.key === "~" || (e.key === "`" && e.shiftKey)) {
        const target = e.target as HTMLElement
        const isInput =
          target.tagName === "INPUT" ||
          target.tagName === "TEXTAREA" ||
          target.isContentEditable
        if (isInput) return

        e.preventDefault()
        onOpenChange(true)
      }
    }

    document.addEventListener("keydown", handleKeyDown)
    return () => document.removeEventListener("keydown", handleKeyDown)
  }, [onOpenChange])

  const runAndClose = (action: ProcessAction, processId: string) => {
    onAction(action, processId)
    onOpenChange(false)
  }

  const openTerminalAndClose = (processId: string) => {
    onOpenTerminal(processId)
    onOpenChange(false)
  }

  return (
    <CommandDialog
      open={open}
      onOpenChange={onOpenChange}
      title="Command Palette"
      description="Search processes and run actions"
    >
      <CommandInput placeholder="Search processes…" />
      <CommandList>
        <CommandEmpty>No results found.</CommandEmpty>

        {processes.map((process, index) => {
          const label = process.name || process.id
          const stopKey = `stop:${process.id}`
          const restartKey = `restart:${process.id}`
          const removeKey = `remove:${process.id}`

          return (
            <CommandGroup key={process.id} heading={label}>
              <CommandItem
                value={`${label} terminal open terminal ${index}`}
                onSelect={() => openTerminalAndClose(process.id)}
              >
                <SquareTerminal className="text-slate-400" />
                <span>Open Terminal</span>
              </CommandItem>
              <CommandItem
                value={`${label} restart ${index}`}
                disabled={busyActionKey === restartKey}
                onSelect={() => runAndClose("restart", process.id)}
              >
                <RotateCcw className="text-sky-400" />
                <span>Restart</span>
              </CommandItem>
              <CommandItem
                value={`${label} stop ${index}`}
                disabled={busyActionKey === stopKey}
                onSelect={() => runAndClose("stop", process.id)}
              >
                <Square className="text-amber-400" />
                <span>Stop</span>
              </CommandItem>
              <CommandItem
                value={`${label} remove delete ${index}`}
                disabled={busyActionKey === removeKey}
                onSelect={() => runAndClose("remove", process.id)}
              >
                <Trash2 className="text-red-400" />
                <span>Remove</span>
              </CommandItem>
              {index < processes.length - 1 && <CommandSeparator />}
            </CommandGroup>
          )
        })}
      </CommandList>
    </CommandDialog>
  )
}
