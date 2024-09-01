import { QueryClient, QueryClientProvider } from "react-query";
import { createTheme, MantineProvider } from "@mantine/core";
import "@mantine/core/styles.css";
import "./App.css";

const queryClient = new QueryClient();

const theme = createTheme({
    /** Put your mantine theme override here */
});

// pages
import Home from "./pages/Home";

function App() {
    return (
        <MantineProvider theme={theme} defaultColorScheme="dark">
            <QueryClientProvider client={queryClient}>
                <Home />
            </QueryClientProvider>
        </MantineProvider>
    );
}

export default App;
