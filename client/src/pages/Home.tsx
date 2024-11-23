import { useState } from "react";
import { useQuery } from "react-query";
import axios from "axios";
import {
    Container,
    Flex,
    Button,
    Text,
    Title,
    Image,
    Anchor,
    Tooltip,
    Modal,
    ActionIcon,
    MultiSelect,
    Card,
    Rating,
    Skeleton,
    LoadingOverlay,
} from "@mantine/core";
import { IconAdjustmentsHorizontal, IconAi } from "@tabler/icons-react";
import { useNavigate } from "react-router-dom";
import { useDisclosure } from "@mantine/hooks";
import { useLocation } from "../LocationContext";

// Interfaces
interface Restaurant {
    id: string;
    name: string;
    chain_name: string;
    hero_image: string;
    budget: number;
    rating: number;
    review_number: number;
    redirection_url: string;
    weight: number;
}

const API_ROUTE = {
    // baseURL: "http://what-to-eat.jraylin.top",
    baseURL: "http://localhost:3000",
    endpoints: {
        getRandomRestaurant: "/api/v1/picker/random",
    },
};

const useRandomRestaurant = (
    latitude: number | undefined,
    longitude: number | undefined,
    cuisineTypes: string[]
) => {
    return useQuery({
        queryKey: ["randomRestaurant", latitude, longitude, cuisineTypes],
        queryFn: async () => {
            if (
                typeof latitude === "undefined" ||
                typeof longitude === "undefined"
            ) {
                throw new Error("Location is required");
            }

            const response = await axios.get(
                `${API_ROUTE.baseURL}${API_ROUTE.endpoints.getRandomRestaurant}`,
                {
                    params: {
                        latitude,
                        longitude,
                        cuisineTypes: cuisineTypes.join(","),
                    },
                }
            );
            return response.data.restaurants[0] as Restaurant;
        },
        enabled: false,
        staleTime: 0,
    });
};

function Home() {
    const navigate = useNavigate();
    const [imageError, setImageError] = useState<boolean>(false);
    const [opened, { open, close }] = useDisclosure(false);
    const [foodType, setFoodType] = useState<string[]>([]);

    const {
        location,
        cuisines,
        isLoading: isInitializing,
        error: locationError,
    } = useLocation();

    const {
        data: restaurant,
        isLoading: isLoadingRestaurant,
        error: restaurantError,
        refetch: getRandomRestaurant,
    } = useRandomRestaurant(location?.latitude, location?.longitude, foodType);

    const error = locationError || restaurantError;

    const SkeletonCard = () => (
        <Card
            className="restaurantCard"
            shadow="sm"
            padding="lg"
            radius="md"
            withBorder
        >
            <Card.Section>
                <Skeleton className="skeleton-image" height={200} />
            </Card.Section>
            <Skeleton height={24} mt="md" width="70%" />
            <Skeleton height={20} mt="sm" width="40%" />
        </Card>
    );

    const LoadedCard = ({ restaurant }: { restaurant: Restaurant }) => (
        <Card
            className="restaurantCard"
            shadow="sm"
            padding="lg"
            radius="md"
            withBorder
        >
            <Anchor href={restaurant.redirection_url} target="_blank">
                <Card.Section>
                    <div className="image-container">
                        <div className="image-content">
                            <Image
                                className="heroImage"
                                src={
                                    imageError
                                        ? "ghost.svg"
                                        : restaurant.hero_image ||
                                          "https://placehold.co/600x400"
                                }
                                alt="Restaurant Image"
                                onLoad={() => {
                                    setImageError(false);
                                }}
                                onError={() => {
                                    setImageError(true);
                                }}
                            />
                        </div>
                    </div>
                </Card.Section>

                <Text fz="lg" fw={500} mt="md">
                    {restaurant.name}
                </Text>
                <Tooltip
                    label={`評論數 ${restaurant.review_number}`}
                    position="bottom"
                >
                    <Rating value={restaurant.rating} fractions={5} readOnly />
                </Tooltip>
            </Anchor>
        </Card>
    );

    if (error) {
        return (
            <Container size="sm" mt={16}>
                <Text color="red">Error: {error.message}</Text>
            </Container>
        );
    }

    return (
        <div className="app-container">
            <div className="content-wraper">
                <LoadingOverlay visible={isInitializing} />

                <Modal
                    opened={opened}
                    onClose={close}
                    title="Settings"
                    yOffset="16vh"
                    overlayProps={{
                        backgroundOpacity: 0.25,
                        blur: 20,
                    }}
                >
                    <MultiSelect
                        label="食物類別"
                        value={foodType}
                        data={
                            cuisines?.map((cuisine) => ({
                                label: cuisine.title,
                                value: cuisine.id.toString(),
                            })) || []
                        }
                        onChange={setFoodType}
                        disabled={isInitializing}
                    />
                </Modal>

                <Container size="100%" mt={16}>
                    <Tooltip label="Settings" position="bottom">
                        <ActionIcon
                            aria-label="Settings"
                            size="xl"
                            variant="transparent"
                            onClick={open}
                            style={{ position: "absolute", top: 16, left: 16 }}
                            disabled={isInitializing}
                        >
                            <IconAdjustmentsHorizontal
                                size={32}
                                stroke={1.5}
                                color="#ffffff"
                            />
                        </ActionIcon>
                    </Tooltip>

                    <Flex
                        mih={50}
                        gap="md"
                        justify="center"
                        align="center"
                        direction="column"
                        wrap="wrap"
                    >
                        <Title>What to eat</Title>
                        <Button
                            onClick={() => getRandomRestaurant()}
                            disabled={isInitializing || isLoadingRestaurant}
                            loading={isLoadingRestaurant}
                        >
                            幫我選
                        </Button>

                        <div className="card-container">
                            {isLoadingRestaurant ? (
                                <SkeletonCard />
                            ) : restaurant ? (
                                <LoadedCard restaurant={restaurant} />
                            ) : null}
                        </div>
                    </Flex>

                    <Tooltip label="AI suggestion" position="bottom">
                        <ActionIcon
                            aria-label="Settings"
                            size="xl"
                            onClick={() => navigate("/ai")}
                            variant="transparent"
                            style={{ position: "absolute", top: 16, right: 16 }}
                            disabled={isInitializing}
                        >
                            <IconAi size={48} stroke={1.5} color="#ffffff" />
                        </ActionIcon>
                    </Tooltip>
                </Container>
            </div>

            <Text className="Footer">
                {"Copyright © CHUN-JUI, LIN "}
                {new Date().getFullYear()}. All Rights Reserved.
            </Text>
        </div>
    );
}

export default Home;
