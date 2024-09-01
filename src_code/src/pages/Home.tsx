import { useState, useEffect } from "react";
import axios from "axios";
import { useQuery } from "react-query";
import { useDisclosure } from "@mantine/hooks";
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
} from "@mantine/core";
import { IconAdjustmentsHorizontal } from "@tabler/icons-react";

// Interfaces
interface GeoLocationInfo {
    coords: {
        latitude: number;
        longitude: number;
    };
    timestamp: number;
}

interface Location {
    latitude: number;
    longitude: number;
}

interface ChainInfo {
    code: string;
    name: string;
    main_vendor_code: string;
    url_key: string;
}

interface RawRestaurant {
    id: string;
    name: string;
    chain: ChainInfo;
    hero_image: string;
    budget: number;
    rating: number;
    review_number: number;
    redirection_url: string;
    metadata: {
        is_delivery_available: boolean;
        is_pickup_available: boolean;
    };
}

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

interface Cuisines {
    id: number;
    title: string;
    count: number;
}

function Home() {
    const [imageError, setImageError] = useState<boolean>(false);
    const [imageLoading, setImageLoading] = useState<boolean>(false);
    const [isLotteryClicked, setIsLotteryClicked] = useState<boolean>(false);
    const [opened, { open, close }] = useDisclosure(false);
    const [availableCuisines, setAvailableCuisines] = useState<Cuisines[]>([]);
    const [foodType, setFoodType] = useState<string[]>([]);
    const [usrLocation, setUsrLocation] = useState<Location>({
        latitude: 0,
        longitude: 0,
    });
    const [lotteryResult, setLotteryResult] = useState<Restaurant | null>(null);

    const foodpandaAPI = {
        url: "https://disco.deliveryhero.io/listing/api/v1/pandora/vendors",
        headers: { "x-disco-client-id": "web" },
        params: {
            country: "tw",
            latitude: usrLocation.latitude,
            longitude: usrLocation.longitude,
            language_id: 6,
            include: "characteristics",
            dynamic_pricing: 0,
            configuration: "Original",
            budgets: "",
            cuisine: foodType.join(","),
            sort: "",
            food_characteristic: "",
            use_free_delivery_label: false,
            vertical: "restaurants",
            limit: 999999,
            offset: 0,
            customer_type: "regular",
        },
    };

    function promoteAlgorithm(rating: number, review_number: number): number {
        const maxReview = 200;
        const offset = 50;
        const weight = Math.min(1, maxReview / review_number);
        const point =
            (review_number * rating + offset * weight) /
            (review_number + offset * weight);
        return point;
    }

    const getNearRestaurant = async (): Promise<Restaurant[]> => {
        const response = await axios.get(foodpandaAPI.url, {
            headers: foodpandaAPI.headers,
            params: foodpandaAPI.params,
        });

        if (response.status === 200) {
            const cuisines: Cuisines[] =
                response.data.data.aggregations.cuisines;
            setAvailableCuisines(cuisines);
            const rawRestaurants: RawRestaurant[] = response.data.data.items;
            // console.log(rawRestaurants);

            const processedRestaurants: Restaurant[] = rawRestaurants
                .filter(
                    (restaurant: RawRestaurant) =>
                        restaurant.metadata.is_delivery_available === true
                )
                .map((restaurant: RawRestaurant) => ({
                    id: restaurant.id,
                    name: restaurant.name,
                    chain_name: restaurant.chain.name,
                    hero_image: restaurant.hero_image,
                    budget: restaurant.budget,
                    rating: restaurant.rating,
                    review_number: restaurant.review_number,
                    redirection_url: restaurant.redirection_url,
                    weight: promoteAlgorithm(
                        restaurant.rating,
                        restaurant.review_number
                    ),
                }));
            return processedRestaurants;
        } else {
            throw new Error("Error fetching restaurants");
        }
    };

    const {
        data: restaurants,
        isLoading,
        error,
        refetch,
    } = useQuery({
        queryKey: ["foodpandaAPI", foodType],
        queryFn: getNearRestaurant,
        enabled: usrLocation.latitude !== 0 && usrLocation.longitude !== 0,
    });

    function getUsrLocation() {
        function usrLocation(position: GeoLocationInfo) {
            const usrLocationInfo: Location = {
                latitude: position.coords.latitude,
                longitude: position.coords.longitude,
            };

            setUsrLocation(usrLocationInfo);
        }

        function error() {
            console.log("Unable to retrieve your location");
        }

        navigator.geolocation.getCurrentPosition(usrLocation, error);
    }

    function lottery(restaurants: Restaurant[]): Restaurant | undefined {
        const totalWeight = restaurants.reduce(
            (acc: number, restaurant: Restaurant) => acc + restaurant.weight,
            0
        );

        const randomNumber = Math.random() * totalWeight;
        let accumulatedWeight = 0;

        for (const restaurant of restaurants) {
            accumulatedWeight += restaurant.weight;
            if (randomNumber <= accumulatedWeight) {
                return restaurant;
            }
        }

        return undefined;
    }

    const handleLottery = () => {
        setIsLotteryClicked(true);
        if (restaurants) {
            const result = lottery(restaurants);
            if (result) {
                setLotteryResult(result);
                setImageLoading(true);
            }
        }
    };

    // Reset image states when a new restaurant is selected
    useEffect(() => {
        if (lotteryResult) {
            setImageLoading(true);
            setImageError(false);
        }
    }, [lotteryResult]);

    useEffect(() => {
        getUsrLocation();
    }, []);

    useEffect(() => {
        if (foodType.length > 0) {
            refetch();
            setLotteryResult(null);
        }
    }, [foodType, refetch]);

    if (error) return <p>An error occurred: {(error as Error).message}</p>;

    const SkeletonCard = () => (
        <Card
            className="restaurantCard"
            shadow="sm"
            padding="lg"
            radius="md"
            withBorder
        >
            <Card.Section>
                <Skeleton className="skeleton-image" />
            </Card.Section>
            <Skeleton
                height={24}
                mt="md"
                width="70%"
                className="skeleton-text"
            />
            <Skeleton
                height={20}
                mt="sm"
                width="40%"
                className="skeleton-rating"
            />
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
                                    setImageLoading(false);
                                    setImageError(false);
                                }}
                                onError={() => {
                                    setImageLoading(false);
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

    return (
        <div className="app-container">
            <div className="content-wraper">
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
                        data={availableCuisines.map((cuisine) => ({
                            label: cuisine.title,
                            value: cuisine.id.toString(),
                        }))}
                        onChange={(selectedItems) => {
                            setFoodType(selectedItems);
                        }}
                    />
                </Modal>
                <Container size={"100%"} mt={16}>
                    <Tooltip label="Settings" position="bottom">
                        <ActionIcon
                            aria-label="Settings"
                            size="xl"
                            variant="transparent"
                            onClick={open}
                            style={{ position: "absolute", top: 16, left: 16 }}
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
                            onClick={handleLottery}
                            disabled={isLoading || !restaurants}
                        >
                            幫我選
                        </Button>

                        <div className="card-container">
                            {isLotteryClicked &&
                                (!imageLoading || lotteryResult ? (
                                    <LoadedCard restaurant={lotteryResult!} />
                                ) : (
                                    <SkeletonCard />
                                ))}
                        </div>
                    </Flex>
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
