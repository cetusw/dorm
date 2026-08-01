import { Alert, Box, Stack, Title } from '@mantine/core'

type Props = {
    title: string
}

export function ResidentSectionPlaceholderPage({ title }: Props) {
    return (
        <Box px={{ base: 'md', md: 'xl' }} py="xl">
            <Stack gap="lg" maw={1240} mx="auto">
                <Title order={1}>{title}</Title>
                <Alert color="gray">
                    Раздел пока не подключен.
                </Alert>
            </Stack>
        </Box>
    )
}
