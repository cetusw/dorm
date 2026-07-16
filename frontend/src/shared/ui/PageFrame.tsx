import type {ReactNode} from 'react'

import {Alert, Box, Group, Stack, Title} from '@mantine/core'

type Props = {
    title: string
    subtitle?: string
    titleActions?: ReactNode
    error?: string | null
    notice?: string | null
    controls?: ReactNode
    analytics?: ReactNode
    children: ReactNode
}

export function PageFrame({
                              title,
                              subtitle,
                              titleActions,
                              error,
                              notice,
                              controls,
                              analytics,
                              children,
                          }: Props) {
    return (
        <Box px={{base: 'md', md: 'xl'}} py="xl">
            <Stack gap="lg" maw={1240} mx="auto">
                {error && (
                    <Alert color="red" title="Ошибка">
                        {error}
                    </Alert>
                )}

                {notice && (
                    <Alert color="gray">
                        {notice}
                    </Alert>
                )}

                <Group justify="space-between" align="flex-start" gap="md" wrap="wrap">
                    <Title order={1}>
                        <Box component="span">
                            {title}
                        </Box>

                        {subtitle && (
                            <>
                                <Box
                                    component="span"
                                    visibleFrom="sm"
                                >
                                    {' · '}
                                </Box>

                                <Box
                                    component="span"
                                    display={{base: 'block', sm: 'inline'}}
                                    mt={{base: 4, sm: 0}}
                                    style={{
                                        whiteSpace: 'nowrap',
                                    }}
                                >
                                    {subtitle}
                                </Box>
                            </>
                        )}
                    </Title>

                    {titleActions}
                </Group>

                {(analytics || controls) && (
                    <Box
                        pos="sticky"
                        top={60}
                        style={{
                            zIndex: 10,
                        }}
                    >
                        <Stack
                            gap="lg"
                            py="sm"
                            bg="#f5f7fb"
                        >
                            {analytics}

                            {controls}
                        </Stack>
                    </Box>
                )}

                {children}
            </Stack>
        </Box>
    )
}
