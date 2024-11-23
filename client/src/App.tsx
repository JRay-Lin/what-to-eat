import { QueryClient, QueryClientProvider } from "react-query";
import { createTheme, MantineProvider } from "@mantine/core";
import { Route, Routes, BrowserRouter } from "react-router-dom";
import { LocationProvider } from "./LocationContext";
import "@mantine/core/styles.css";
import "./App.css";

const queryClient = new QueryClient();

const theme = createTheme({
    /** Put your mantine theme override here */
});

// pages
import Home from "./pages/Home";
import Ai from "./pages/Ai";

function Router() {
    return (
        <div>
            <BrowserRouter>
                <Routes>
                    <Route path="/" element={<Home />} />
                    <Route path="/ai" element={<Ai />} />
                </Routes>
            </BrowserRouter>
        </div>
    );
}

function App() {
    return (
        <MantineProvider theme={theme} defaultColorScheme="dark">
            <QueryClientProvider client={queryClient}>
                <LocationProvider>
                    <Router />
                </LocationProvider>
            </QueryClientProvider>
        </MantineProvider>
    );
}

export default App;
