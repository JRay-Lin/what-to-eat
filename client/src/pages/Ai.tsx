import {
    Container,
    Flex,
    Button,
    Text,
    Title,
    Tooltip,
    ActionIcon,
    TextInput,
    ScrollArea,
} from "@mantine/core";
import { IconDice6, IconSend } from "@tabler/icons-react";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

function Ai() {
    const navigate = useNavigate();
    const [messages, setMessages] = useState([
        { sender: "User", content: "Hello! What should I eat today?" },
        { sender: "AI", content: "How about trying some Italian cuisine?" },
    ]);
    const [newMessage, setNewMessage] = useState("");

    const handleSendMessage = () => {
        if (newMessage.trim()) {
            setMessages([...messages, { sender: "User", content: newMessage }]);
            setNewMessage(""); // Clear input after sending
        }
    };

    return (
        <div className="app-container">
            <div className="content-wraper">
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

                        {/* Chat Room Interface */}
                        <ScrollArea
                            style={{
                                height: 300,
                                width: "100%",
                                border: "1px solid #ccc",
                                borderRadius: "8px",
                                padding: "10px",
                            }}
                        >
                            <div className="chat-messages">
                                {messages.map((message, index) => (
                                    <Text
                                        key={index}
                                        align={
                                            message.sender === "User"
                                                ? "right"
                                                : "left"
                                        }
                                        style={{ margin: "5px 0" }}
                                    >
                                        <strong>{message.sender}:</strong>{" "}
                                        {message.content}
                                    </Text>
                                ))}
                            </div>
                        </ScrollArea>

                        <Flex
                            mt={8}
                            gap="sm"
                            align="center"
                            style={{ width: "100%" }}
                        >
                            <TextInput
                                placeholder="Type your message..."
                                value={newMessage}
                                onChange={(e) =>
                                    setNewMessage(e.currentTarget.value)
                                }
                                style={{ flex: 1 }}
                            />
                            <Button
                                onClick={handleSendMessage}
                                variant="outline"
                                color="blue"
                            >
                                <IconSend size={18} />
                            </Button>
                        </Flex>
                    </Flex>

                    <Tooltip label="Settings" position="bottom">
                        <ActionIcon
                            aria-label="Settings"
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
