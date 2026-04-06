import { ProcessManagerPage } from "@/features/process-manager/ProcessManagerPage"
import { ThemeProvider } from "./components/ui/theme-provider"

function App() {
  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <ProcessManagerPage />
    </ThemeProvider>
  )
}

export default App
