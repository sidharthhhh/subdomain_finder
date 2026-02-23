import axios from "axios";

const rawUrl = import.meta.env.VITE_API_URL || "http://localhost:8080";
const API_BASE_URL = rawUrl.replace(/\/+$/, "");

export interface Subdomain {
    domain: string;
    source: "CT_LOG" | "BRUTE_FORCE" | "WILDCARD";
    ip?: string;
    found_at: string;
}

export interface ScanResponse {
    subdomains: string[];
    count: number;
}

export const api = {
    scan: async (domain: string): Promise<ScanResponse> => {
        const response = await axios.post<ScanResponse>(`${API_BASE_URL}/scan`, {
            domain,
        }, {
            timeout: 300000, // 5 minutes timeout for massive scans (Google, etc)
        });
        return response.data;
    },
};
