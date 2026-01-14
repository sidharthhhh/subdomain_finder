import axios from "axios";

const API_BASE_URL = "http://localhost:8080";

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
        });
        return response.data;
    },
};
