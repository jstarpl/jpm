import { Separator } from "@/components/ui/separator"
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
} from "@/components/ui/sheet"
import type { Process } from "./types"

type ProcessDetailsSheetProps = {
  process: Process | null
  onClose: () => void
}

export function ProcessDetailsSheet({ process, onClose }: ProcessDetailsSheetProps) {
  return (
    <Sheet open={Boolean(process)} onOpenChange={(open) => !open && onClose()}>
      <SheetContent side="right" className="w-[90vw] border-slate-100/10 bg-slate-950 text-slate-100 sm:max-w-xl">
        <SheetHeader>
          <SheetTitle>Process Details</SheetTitle>
          <SheetDescription className="text-slate-300">
            {process ? `${process.name || "(unnamed)"} (${process.id})` : "No process selected"}
          </SheetDescription>
        </SheetHeader>

        {process && (
          <div className="grid gap-4 px-4 pb-6">
            <div className="rounded-md border border-slate-100/10 bg-slate-900/40 p-3">
              <p className="text-xs uppercase tracking-wide text-slate-400">Executable</p>
              <p className="mt-1 break-all text-sm">{process.exec}</p>
              <p className="mt-3 text-xs uppercase tracking-wide text-slate-400">Working Directory</p>
              <p className="mt-1 break-all text-sm">{process.cwd || "-"}</p>
            </div>

            <Separator className="bg-slate-100/10" />

            <div>
              <h3 className="text-sm font-semibold">Arguments ({process.args.length})</h3>
              <ul className="mt-2 max-h-40 space-y-2 overflow-auto rounded-md border border-slate-100/10 bg-slate-900/40 p-3 text-sm">
                {process.args.length === 0 && <li className="text-slate-400">No arguments</li>}
                {process.args.map((arg, index) => (
                  <li key={`${process.id}-arg-${index}`} className="break-all text-slate-200">
                    {arg}
                  </li>
                ))}
              </ul>
            </div>

            <div>
              <h3 className="text-sm font-semibold">Environment ({process.env.length})</h3>
              <ul className="mt-2 max-h-64 space-y-2 overflow-auto rounded-md border border-slate-100/10 bg-slate-900/40 p-3 text-sm">
                {process.env.length === 0 && <li className="text-slate-400">No environment variables</li>}
                {process.env.map((envVar, index) => (
                  <li key={`${process.id}-env-${index}`} className="break-all text-slate-200">
                    {envVar}
                  </li>
                ))}
              </ul>
            </div>
          </div>
        )}
      </SheetContent>
    </Sheet>
  )
}
