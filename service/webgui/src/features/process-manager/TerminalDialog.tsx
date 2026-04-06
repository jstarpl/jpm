import { useLayoutEffect, useState } from "react"
import * as DialogPrimitive from "@radix-ui/react-dialog"
import { FitAddon } from "@xterm/addon-fit"
import { Terminal } from "@xterm/xterm"
import "@xterm/xterm/css/xterm.css"
import { Button } from "@/components/ui/button"
import { createProcessManagerApi } from "./api"

type TerminalDialogProps = {
  processId: string | null
  token: string
  onOpenChange: (open: boolean) => void
}

export function TerminalDialog({ processId, token, onOpenChange }: TerminalDialogProps) {
  const [containerRef, setContainerRef] = useState<HTMLDivElement | null>(null)
  const [connectionError, setConnectionError] = useState<string | null>(null)

  useLayoutEffect(() => {
    setConnectionError(null)

    if (processId === null || !containerRef) {
      return
    }

    const api = createProcessManagerApi(token)
    const term = new Terminal({
      fontFamily: "Consolas, 'Courier New', monospace",
      fontSize: 14,
      convertEol: true,
      cursorBlink: true,
      theme: {
        background: "#020617",
        foreground: "#dbeafe",
      },
    })

    const fitAddon = new FitAddon()
    term.loadAddon(fitAddon)
    term.open(containerRef)
    fitAddon.fit()

    term.writeln(`Connected to process ${processId}. Type and press Enter to send stdin.`)
    term.write("\r\n")

    const queryToken = token ? `?token=${encodeURIComponent(token)}` : ""
    const stream = new EventSource(`/api/processes/${processId}/stdouterr${queryToken}`)

    const writeEvent = (event: MessageEvent<string>) => {
      term.write(event.data + "\r\n")
    }

    stream.addEventListener("stdout", writeEvent as EventListener)
    stream.addEventListener("stderr", writeEvent as EventListener)
    stream.onerror = () => {
      setConnectionError("Terminal stream disconnected")
    }

    let stdinBuffer = ""

    const sendStdin = async (value: string) => {
      if (!value) {
        return
      }

      try {
        await api.sendStdin(processId, value)
      } catch {
        setConnectionError("Failed to send stdin data")
      }
    }

    const disposable = term.onData((chunk) => {
      for (const char of Array.from(chunk)) {
        if (char === "\r" || char === "\n") {
          const payload = `${stdinBuffer}\n`
          stdinBuffer = ""
          term.write("\r\n")
          void sendStdin(payload)
          term.write("$ ")
          continue
        }

        if (char === "\u007F") {
          if (stdinBuffer.length > 0) {
            stdinBuffer = stdinBuffer.slice(0, -1)
            term.write("\b \b")
          }
          continue
        }

        stdinBuffer += char
        term.write(char)
      }
    })

    const onResize = () => fitAddon.fit()
    window.addEventListener("resize", onResize)

    return () => {
      window.removeEventListener("resize", onResize)
      disposable.dispose()
      stream.close()
      term.dispose()
    }
  }, [processId, token, containerRef])

  return (
    <DialogPrimitive.Root open={processId !== null} onOpenChange={onOpenChange}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Overlay className="fixed inset-0 z-50 bg-black/60 backdrop-blur-sm" />
        <DialogPrimitive.Content className="fixed top-1/2 left-1/2 z-50 flex h-[85vh] w-[95vw] max-w-5xl -translate-x-1/2 -translate-y-1/2 flex-col gap-3 rounded-lg border border-slate-100/10 bg-slate-950 p-4 shadow-xl">
          <div className="flex items-center justify-between gap-2">
            <DialogPrimitive.Title className="text-base font-semibold text-slate-100">
              Interactive Terminal {processId ? `- ${processId}` : ""}
            </DialogPrimitive.Title>
            <DialogPrimitive.Close asChild>
              <Button variant="outline" className="border-slate-100/20 bg-slate-900/30">
                Close
              </Button>
            </DialogPrimitive.Close>
          </div>

          {connectionError && <p className="text-sm text-red-300">{connectionError}</p>}

          <div className="h-full overflow-hidden rounded-md border border-slate-100/10 bg-slate-950 p-2">
            <div ref={setContainerRef} className="h-full w-full" />
          </div>
        </DialogPrimitive.Content>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  )
}
