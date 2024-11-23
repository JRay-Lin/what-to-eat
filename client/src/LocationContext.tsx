import { createContext, useContext, ReactNode } from "react";
import { useQuery } from "react-query";
import axios from "axios";

interface Location {
    latitude: number;
    longitude: number;
}

interface Cuisines {
    id: number;
    title: string;
    count: number;
}

interface LocationContextType {
    location: Location | undefined;
    cuisines: Cuisines[] | undefined;
    isLoading: boolean;
    error: Error | null;
}

const LocationContext = createContext<LocationContextType | undefined>(
    undefined
);

const API_ROUTE = {
    baseURL: "http://localhost:3000",
    endpoints: {
        getAvailableCuisines: "/api/v1/cuisines",
    },
};

export function LocationProvider({ children }: { children: ReactNode }) {
    const {
        data: location,
        isLoading: isLoadingLocation,
        error: locationError,
    } = useQuery({
        queryKey: ["geolocation"],
        queryFn: (): Promise<Location> => {
            return new Promise((resolve, reject) => {
                if (!navigator.geolocation) {
                    reject(
                        new Error(
                            "Geolocation is not supported by your browser"
                        )
                    );
                    return;
                }

                navigator.geolocation.getCurrentPosition(
                    (position) => {
                        resolve({
                            latitude: position.coords.latitude,
                            longitude: position.coords.longitude,
                        });
                    },
                    (error) => {
                        reject(
                            new Error(`Geolocation error: ${error.message}`)
                        );
                    }
                );
            });
        },
        retry: 1,
        staleTime: 5 * 60 * 1000,
    });

    const {
        data: cuisines,
        isLoading: isLoadingCuisines,
        error: cuisinesError,
    } = useQuery({
        queryKey: ["cuisines", location],
        queryFn: async () => {
            if (!location) throw new Error("Location is required");

            const response = await axios.get(
                `${API_ROUTE.baseURL}${API_ROUTE.endpoints.getAvailableCuisines}`,
                {
                    params: {
                        latitude: location.latitude,
                        longitude: location.longitude,
                    },
                }
            );
            return response.data.cuisines as Cuisines[];
        },
        enabled: !!location,
        staleTime: 5 * 60 * 1000,
    });

    const isLoading = isLoadingLocation || isLoadingCuisines;
    const error = locationError || cuisinesError || null;

    return (
        <LocationContext.Provider
            value={{ location, cuisines, isLoading, error }}
        >
            {children}
        </LocationContext.Provider>
    );
}

export function useLocation() {
    const context = useContext(LocationContext);
    if (context === undefined) {
        throw new Error("useLocation must be used within a LocationProvider");
    }
    return context;
}
