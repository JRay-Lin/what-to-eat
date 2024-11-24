import {
  Container,
  Flex,
  Grid,
  Button,
  Text,
  Title,
  Tooltip,
  ActionIcon,
  Stepper,
  Group,
  Chip,
  Textarea,
  LoadingOverlay,
  Alert,
  Card,
  Anchor,
  Image,
} from "@mantine/core";
import { IconDice6, IconAlertCircle } from "@tabler/icons-react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useLocation } from "../LocationContext";

interface Category {
  id: number;
  label: string;
}

interface Restaurant {
  id: string;
  code: string;
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
  baseURL: "http://localhost:3000",
  endpoints: {
    getFilteredCategories: "/api/v1/picker/ai/filter-categories",
    getFinalSuggestion: "/api/v1/picker/ai/suggestion",
  },
};

function Ai() {
  const navigate = useNavigate();
  const [isLoading, setIsLoading] = useState(false);
  const [imageError, setImageError] = useState<boolean>(false);

  // User preference
  const [initialPreference, setInitialPreference] = useState("");
  const [additionalDetails, setAdditionalDetails] = useState("");
  const [selectedCategories, setSelectedCategories] = useState<string[]>([]);

  // API response
  const [filteredCategories, setFilteredCategories] = useState<Category[]>([]);
  const [aiRestaurantSuggestion, setAiRestaurantSuggestion] =
    useState<Restaurant | null>(null);

  // Stepper
  const [active, setActive] = useState(0);

  // Geolocation & local cuisine types
  const {
    location,
    cuisines,
    isLoading: isInitializing,
    error: locationError,
  } = useLocation();

  // Available categories from context
  const availableCategories: Category[] =
    cuisines?.map((cuisine) => ({
      id: cuisine.id,
      label: cuisine.title,
    })) || [];

  const handleSelection = (value: string[]) => {
    setSelectedCategories(value);
  };

  const handleNextStep = async () => {
    try {
      setIsLoading(true);

      if (active === 0) {
        // First step: Send initial preference and get filtered categories
        if (!initialPreference.trim()) {
          alert("請輸入想吃的內容");
          return;
        }

        const response = await fetch(
          `${API_ROUTE.baseURL}${API_ROUTE.endpoints.getFilteredCategories}`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({
              userPreference: initialPreference,
              location,
              availableCategories: availableCategories,
            }),
          }
        );

        if (!response.ok) {
          throw new Error("Failed to get category suggestions");
        }

        const data = await response.json();
        // console.log(data);
        setFilteredCategories(data);
        setActive(1);
      } else if (active === 1) {
        // Second step: Validate category selection
        if (selectedCategories.length === 0) {
          alert("請選擇至少一個類別");
          return;
        }
        setActive(2);
      } else if (active === 2) {
        // Final step: Send all data and get final suggestion
        if (!additionalDetails.trim()) {
          alert("請輸入額外需求");
          return;
        }

        const response = await fetch(
          `${API_ROUTE.baseURL}${API_ROUTE.endpoints.getFinalSuggestion}`,
          {
            method: "POST",
            headers: {
              "Content-Type": "application/json",
            },
            body: JSON.stringify({
              initial_preference: initialPreference,
              additional_details: additionalDetails,
              location,
              selectedCategories,
            }),
          }
        );

        if (!response.ok) {
          throw new Error("Failed to get final suggestion");
        }

        const data = await response.json();
        setAiRestaurantSuggestion(data);
        console.log(data);
        setActive(3);
      }
    } catch (error) {
      console.error("Error:", error);
      alert("發生錯誤，請稍後再試");
    } finally {
      setIsLoading(false);
    }
  };

  const prevStep = () => {
    setActive((current) => (current > 0 ? current - 1 : current));
  };

  if (locationError) {
    return (
      <Container size="sm" mt={16}>
        <Alert
          icon={<IconAlertCircle size={16} />}
          title="Error"
          color="red"
          variant="filled"
        >
          {locationError.message}
        </Alert>
      </Container>
    );
  }

  const RestaurantCard = ({ restaurant }: { restaurant: Restaurant }) => (
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
                    : restaurant.hero_image || "https://placehold.co/600x400"
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
      </Anchor>
    </Card>
  );

  return (
    <div className="app-container">
      <div className="content-wraper" style={{ position: "relative" }}>
        <LoadingOverlay
          visible={isLoading || isInitializing}
          // overlayblur={2}
        />
        <Container size={"100%"} mt={16}>
          <Flex
            mih={50}
            gap="md"
            justify="center"
            align="center"
            direction="column"
            wrap="wrap"
          >
            <Title>What to eat</Title>
            <Stepper
              active={active}
              onStepClick={setActive}
              allowNextStepsSelect={false}
            >
              <Stepper.Step label="初步偏好" description="告訴我你想吃什麼口味">
                <Textarea
                  placeholder="想吃什麼呢？例如：想吃辣的、清淡的..."
                  mt={20}
                  value={initialPreference}
                  onChange={(event) =>
                    setInitialPreference(event.currentTarget.value)
                  }
                  styles={() => ({
                    input: {
                      height: "150px",
                    },
                  })}
                />
              </Stepper.Step>
              <Stepper.Step
                label="選擇類別"
                description="根據你的喜好，我們推薦以下類別"
              >
                <Grid mt={20} justify="center">
                  <Chip.Group
                    multiple
                    value={selectedCategories}
                    onChange={handleSelection}
                  >
                    {filteredCategories.map((category) => (
                      <Grid.Col key={category.id} span="content">
                        <Chip value={category.label} size="lg">
                          {category.label}
                        </Chip>
                      </Grid.Col>
                    ))}
                  </Chip.Group>
                </Grid>
              </Stepper.Step>
              <Stepper.Step label="額外需求" description="還有什麼特別要求嗎？">
                <Textarea
                  placeholder="有什麼特別需求嗎？例如：食材，料理方法..."
                  mt={20}
                  value={additionalDetails}
                  onChange={(event) =>
                    setAdditionalDetails(event.currentTarget.value)
                  }
                  styles={() => ({
                    input: {
                      height: "150px",
                    },
                  })}
                />
              </Stepper.Step>
              <Stepper.Completed>
                <div className="suggestion-container">
                  <Flex justify="center" align="center" direction="column" mt={20}>
                    <Title order={2} mb="md">
                      我覺得你應該會喜歡....
                    </Title>
                    {aiRestaurantSuggestion ? (
                      <RestaurantCard restaurant={aiRestaurantSuggestion} />
                    ) : (
                      <Text>沒有找到適合的餐廳</Text>
                    )}
                  </Flex>
                </div>
              </Stepper.Completed>
            </Stepper>

            <Group justify="center" mt="xl">
              <Button
                variant="default"
                onClick={prevStep}
                disabled={active === 0}
              >
                Back
              </Button>
              <Button onClick={handleNextStep} disabled={isLoading}>
                {isLoading
                  ? "Loading..."
                  : active === 2
                  ? "Get Suggestion"
                  : "Next"}
              </Button>
            </Group>
          </Flex>

          <Tooltip label="Back to Home" position="bottom">
            <ActionIcon
              aria-label="Back to Home"
              size="xl"
              onClick={() => navigate("/")}
              variant="transparent"
              style={{ position: "absolute", top: 16, right: 16 }}
            >
              <IconDice6 size={32} stroke={1.5} color="#ffffff" />
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

export default Ai;
