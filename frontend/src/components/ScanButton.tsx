import { Search, Loader2 } from "lucide-react";
import { Button } from "@/components/ui/button";

interface ScanButtonProps {
    onClick: () => void;
    disabled?: boolean;
    loading?: boolean;
}

export function ScanButton({ onClick, disabled, loading }: ScanButtonProps) {
    return (
        <Button
            onClick={onClick}
            disabled={disabled}
            className="h-12 px-8 text-lg shadow-lg hover:shadow-primary/20 transition-all duration-300"
        >
            {loading ? (
                <>
                    <Loader2 className="mr-2 h-5 w-5 animate-spin" />
                    Scanning...
                </>
            ) : (
                <>
                    <Search className="mr-2 h-5 w-5" />
                    Scan Target
                </>
            )}
        </Button>
    );
}
