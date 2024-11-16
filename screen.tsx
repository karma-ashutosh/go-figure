import React, { useState, useRef, useEffect } from 'react'
import { Moon, Sun, ChevronDown, ChevronRight, Play, Clipboard, Save, X } from 'lucide-react'
import { Button } from "@/components/ui/button"
import { Textarea } from "@/components/ui/textarea"
import { Switch } from "@/components/ui/switch"
import { ScrollArea } from "@/components/ui/scroll-area"
import {
    Select,
    SelectContent,
    SelectItem,
    SelectTrigger,
    SelectValue,
} from "@/components/ui/select"

export default function GoFigure() {
    const [darkMode, setDarkMode] = useState(true)
    const [query, setQuery] = useState('')
    const [mode, setMode] = useState('execute')
    const [response, setResponse] = useState([])
    const [history, setHistory] = useState([])
    const [commandOutput, setCommandOutput] = useState('')
    const [isLoading, setIsLoading] = useState(false)
    const textareaRef = useRef(null)

    useEffect(() => {
        const handleKeyDown = (e: KeyboardEvent) => {
            if (e.ctrlKey && e.key === 'Enter') {
                handleSubmit()
            } else if (e.key === 'Escape') {
                setQuery('')
            }
        }
        window.addEventListener('keydown', handleKeyDown)
        return () => window.removeEventListener('keydown', handleKeyDown)
    }, [query])

    const handleSubmit = async () => {
        if (!query.trim()) return;

        setIsLoading(true);
        const callBody = {
            method: "POST",
            headers: {
                "Content-Type": "application/json",
            },
            body: JSON.stringify({ query, mode }),
        }
        console.log(`Will try to invoke with ${JSON.stringify(callBody)} and again with \n ${callBody} and ${JSON.stringify(callBody.body)}`)


        try {
            // Send the query to the backend
            const response = await fetch("http://localhost:8080/api/query", callBody);

            if (!response.ok) {
                throw new Error(`Error: ${response.statusText}`);
            }

            const data = await response.json();

            if (data.error) {
                console.error("API Error:", data.error);
                setResponse([]);
            } else {
                setResponse(data.steps || []);
                setHistory([...history, { query, response: data.steps || [] }]);
            }
        } catch (error) {
            console.error("Error fetching steps:", error);
            setResponse([]);
        } finally {
            setIsLoading(false);
        }
    };

    const executeCommand = (command: string) => {
        setCommandOutput(`Executing: ${command}\n\nOutput: Command executed successfully.`)
    }

    const copyToClipboard = (text: string) => {
        navigator.clipboard.writeText(text)
    }

    const saveToFile = (command: string) => {
        console.log('Saving to file:', command)
    }

    return (
        <div className={`min-h-screen flex flex-col ${darkMode ? 'dark bg-gray-900 text-white' : 'bg-white text-gray-900'}`}>
            <header className="p-4 bg-blue-600 text-white">
                <h1 className="text-3xl font-bold">go-figure</h1>
                <p className="text-xl">Linux Terminal AI</p>
            </header>

            <main className="flex-grow flex overflow-hidden">
                <div className="flex-grow flex flex-col p-4 space-y-4">
                    <div className="flex items-center space-x-4">
                        <Textarea
                            ref={textareaRef}
                            placeholder="Enter your query or error here..."
                            value={query}
                            onChange={(e) => setQuery(e.target.value)}
                            className="flex-grow"
                            rows={4}
                        />
                        <Button onClick={handleSubmit} disabled={isLoading}>
                            {isLoading ? 'Processing...' : 'Submit'}
                        </Button>
                    </div>

                    <div className="flex items-center space-x-4">
                        <Select value={mode} onValueChange={setMode}>
                            <SelectTrigger className="w-[180px]">
                                <SelectValue placeholder="Select mode" />
                            </SelectTrigger>
                            <SelectContent>
                                <SelectItem value="execute">Execute Commands</SelectItem>
                                <SelectItem value="write">Write to File</SelectItem>
                            </SelectContent>
                        </Select>

                        <div className="flex items-center space-x-2 ml-auto">
                            <Sun className="h-4 w-4" />
                            <Switch
                                checked={darkMode}
                                onCheckedChange={setDarkMode}
                            />
                            <Moon className="h-4 w-4" />
                        </div>
                    </div>

                    <ScrollArea className="flex-grow border rounded-md p-4">
                        {response.map((step, index) => (
                            <div key={index} className="mb-4 p-2 border rounded">
                                <h3 className="font-bold">Step {step.step}</h3>
                                <p>{step.description}</p>
                                <p className="text-sm text-gray-500">Reason: {step.reason}</p>
                                {step.command && (
                                    <div className="mt-2 bg-gray-100 dark:bg-gray-800 p-2 rounded">
                                        <code>{step.command}</code>
                                        <div className="mt-2 space-x-2">
                                            {mode === 'execute' && (
                                                <Button size="sm" onClick={() => executeCommand(step.command)}>
                                                    <Play className="h-4 w-4 mr-1" /> Execute
                                                </Button>
                                            )}
                                            <Button size="sm" onClick={() => copyToClipboard(step.command)}>
                                                <Clipboard className="h-4 w-4 mr-1" /> Copy
                                            </Button>
                                            {mode === 'write' && (
                                                <Button size="sm" onClick={() => saveToFile(step.command)}>
                                                    <Save className="h-4 w-4 mr-1" /> Save
                                                </Button>
                                            )}
                                        </div>
                                    </div>
                                )}
                            </div>
                        ))}
                    </ScrollArea>

                    {commandOutput && (
                        <ScrollArea className="h-40 border rounded-md p-4">
                            <pre>{commandOutput}</pre>
                        </ScrollArea>
                    )}
                </div>

                <div className="w-64 p-4 border-l">
                    <h2 className="font-bold mb-2">History</h2>
                    <ScrollArea className="h-full">
                        {history.map((item, index) => (
                            <div key={index} className="mb-2 p-2 border rounded">
                                <details>
                                    <summary className="cursor-pointer">
                                        <ChevronRight className="inline h-4 w-4" />
                                        Query {index + 1}
                                    </summary>
                                    <p className="mt-2 text-sm">{item.query}</p>
                                </details>
                            </div>
                        ))}
                    </ScrollArea>
                </div>
            </main>

            <footer className="p-4 text-center text-sm text-gray-500">
                Use Ctrl+Enter to submit your query. Press Esc to clear the input area.
            </footer>
        </div>
    )
}