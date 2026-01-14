import { Globe } from "lucide-react";
import { Input } from "@/components/ui/input";

interface DomainInputProps {
    value: string;
    onChange: (value: string) => void;
    onEnter: () => void;
    disabled?: boolean;
}

export function DomainInput({ value, onChange, onEnter, disabled }: DomainInputProps) {
    return (
        <div className="relative flex-1">
            <Globe className="absolute left-3 top-2.5 h-5 w-5 text-muted-foreground" />
            <Input
                className="pl-10 h-12 text-lg bg-background/50 border-white/10 focus:border-primary/50 transition-all font-mono"
                placeholder="example.com"
                value={value}
                onChange={(e) => onChange(e.target.value)}
                onKeyDown={(e) => e.key === "Enter" && onEnter()}
                disabled={disabled}
            />
        </div>
    );
}
