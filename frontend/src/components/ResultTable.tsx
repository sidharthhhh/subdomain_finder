import { Download } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

interface ResultTableProps {
    results: string[];
    domain: string;
    onExport: () => void;
}

export function ResultTable({ results, domain, onExport }: ResultTableProps) {
    if (results.length === 0) return null;

    return (
        <Card className="glass-card border-none animate-in fade-in slide-in-from-bottom-4 duration-500">
            <CardHeader className="flex flex-row items-center justify-between space-y-0 pb-4">
                <div>
                    <CardTitle className="text-2xl">Results Found</CardTitle>
                    <CardDescription className="text-base mt-1">
                        Discovered {results.length} unique subdomains for{" "}
                        <span className="text-primary font-mono">{domain}</span>
                    </CardDescription>
                </div>
                <Button
                    variant="outline"
                    onClick={onExport}
                    className="border-primary/20 hover:bg-primary/10"
                >
                    <Download className="mr-2 h-4 w-4" />
                    Export JSON
                </Button>
            </CardHeader>
            <CardContent>
                <div className="rounded-md border border-white/10 bg-black/20 overflow-hidden">
                    <div className="max-h-[500px] overflow-y-auto custom-scrollbar">
                        <table className="w-full text-left">
                            <thead className="bg-white/5 sticky top-0 backdrop-blur-md">
                                <tr>
                                    <th className="p-4 font-medium text-muted-foreground">#</th>
                                    <th className="p-4 font-medium text-muted-foreground w-full">
                                        Subdomain
                                    </th>
                                </tr>
                            </thead>
                            <tbody className="divide-y divide-white/5">
                                {results.map((sub, index) => (
                                    <tr
                                        key={index}
                                        className="hover:bg-white/5 transition-colors"
                                    >
                                        <td className="p-4 text-muted-foreground font-mono text-sm w-16">
                                            {index + 1}
                                        </td>
                                        <td className="p-4 font-mono text-sm text-foreground/90">
                                            {sub}
                                        </td>
                                    </tr>
                                ))}
                            </tbody>
                        </table>
                    </div>
                </div>
            </CardContent>
        </Card>
    );
}
