import { Button } from "@/components/ui/button"
import { ThemeProvider } from "./components/ui/theme-provider"

function App() {
  return (
    <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
      <div className="flex min-h-svh flex-col items-center justify-center">
        <img src="/icon.svg" width="300px" />
        <Button variant="default">Click me</Button>
      </div>
    </ThemeProvider>
  )
}

export default App
