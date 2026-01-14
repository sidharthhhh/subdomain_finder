import { useState } from "react";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { api } from "@/lib/api";
import { ShieldCheck, AlertCircle } from "lucide-react";
import { DomainInput } from "@/components/DomainInput";
import { ScanButton } from "@/components/ScanButton";
import { ResultTable } from "@/components/ResultTable";

function App() {
  const [domain, setDomain] = useState("");
  const [loading, setLoading] = useState(false);
  const [results, setResults] = useState<string[]>([]);
  const [error, setError] = useState<string | null>(null);

  const handleScan = async () => {
    if (!domain) return;
    setLoading(true);
    setError(null);
    setResults([]);
    try {
      const data = await api.scan(domain);
      setResults(data.subdomains);
    } catch (err) {
      setError("Failed to scan domain. Please ensure backend is running.");
      console.error(err);
    } finally {
      setLoading(false);
    }
  };

  const handleExport = () => {
    const blob = new Blob([JSON.stringify({ domain, subdomains: results }, null, 2)], {
      type: "application/json",
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `${domain}-subdomains.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  return (
    <div className="container mx-auto px-4 py-12 max-w-5xl">
      <div className="flex flex-col items-center justify-center text-center space-y-6 mb-12">
        <div className="p-3 rounded-2xl bg-primary/10 border border-primary/20 backdrop-blur-sm">
          <ShieldCheck className="w-12 h-12 text-primary animate-pulse" />
        </div>
        <h1 className="text-4xl md:text-6xl font-bold tracking-tight bg-clip-text text-transparent bg-gradient-to-r from-white to-white/60">
          Subdomain Finder
        </h1>
        <p className="text-lg text-muted-foreground max-w-2xl">
          Advanced reconnaissance tool for uncovering subdomains.
          Powered by high-performance Go backend and Certificate Transparency logs.
        </p>
      </div>

      <div className="grid gap-8">
        <Card className="glass-card border-none">
          <CardHeader>
            <CardTitle>Start New Scan</CardTitle>
            <CardDescription>Enter a domain name to begin the discovery process.</CardDescription>
          </CardHeader>
          <CardContent>
            <div className="flex flex-col md:flex-row gap-4">
              <DomainInput
                value={domain}
                onChange={setDomain}
                onEnter={handleScan}
                disabled={loading}
              />
              <ScanButton
                onClick={handleScan}
                disabled={loading || !domain}
                loading={loading}
              />
            </div>
          </CardContent>
          {error && (
            <CardFooter className="text-destructive font-medium bg-destructive/10 p-4 rounded-b-lg flex items-center gap-2">
              <AlertCircle className="w-5 h-5" />
              {error}
            </CardFooter>
          )}
        </Card>

        <ResultTable
          results={results}
          domain={domain}
          onExport={handleExport}
        />
      </div>
    </div>
  );
}

export default App;
