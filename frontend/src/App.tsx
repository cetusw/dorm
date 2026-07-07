import {Badge, Button, Card, Container, Group, Stack, Text, Title} from '@mantine/core'

function App() {
    return (
        <Container size="sm" py="xl">
            <Stack gap="md">
                <Title order={1}>Сервис уборок</Title>

                <Card withBorder radius="lg" padding="lg">
                    <Group justify="space-between" mb="sm">
                        <Title order={2}>Текущая неделя</Title>
                        <Badge color="green">Дежурство активно</Badge>
                    </Group>

                    <Text c="dimmed" mb="lg">
                        Здесь будут отображаться задачи, которые можно взять, вернуть или отметить выполненными.
                    </Text>

                    <Group>
                        <Button>Мои задачи</Button>
                        <Button variant="light">Свободные задачи</Button>
                    </Group>
                </Card>
            </Stack>
        </Container>
    )
}

export default App