import {useState} from 'react'

import {Alert, Button, Paper, PasswordInput, Stack, TextInput, Title} from '@mantine/core'
import {useForm} from '@mantine/form'

import {ApiError} from '../../../shared/api/ApiError'
import {loginResident} from '../api/authApi'

type LoginFormValues = {
    login: string
    password: string
}

export function LoginForm() {
    const [submitError, setSubmitError] = useState<string | null>(null)
    const [submitting, setSubmitting] = useState(false)

    const form = useForm<LoginFormValues>({
        mode: 'controlled',
        initialValues: {
            login: '',
            password: '',
        },
        validate: {
            login: (value) => (value.trim().length === 0 ? 'Поле не может быть пустым' : null),
            password: (value) => (value.trim().length === 0 ? 'Поле не может быть пустым' : null),
        },
    })

    return (
        <Paper
            radius="lg"
            p={{base: 'lg', sm: 'xl'}}
            shadow="sm"
            withBorder
            w="100%"
            maw={420}
            bg="rgba(255, 255, 255, 0.96)"
        >
            <Title order={2} ta="center">
                Добро пожаловать
            </Title>

            <form
                onSubmit={form.onSubmit(async (values) => {
                    setSubmitError(null)
                    setSubmitting(true)

                    try {
                        const response = await loginResident(values.login, values.password)
                        window.location.assign(response.redirect_url)
                    } catch (error) {
                        if (error instanceof ApiError && error.status === 401) {
                            setSubmitError('Логин или пароль введены некорректно')
                        } else if (error instanceof ApiError) {
                            setSubmitError(error.message)
                        } else {
                            setSubmitError('Не удалось выполнить вход')
                        }
                    } finally {
                        setSubmitting(false)
                    }
                })}
            >
                <Stack gap="md" mt="xl">
                    {submitError && (
                        <Alert color="red" variant="light">
                            {submitError}
                        </Alert>
                    )}

                    <TextInput
                        size="md"
                        withAsterisk
                        label="Логин"
                        placeholder="Введите логин"
                        key={form.key('login')}
                        {...form.getInputProps('login')}
                    />

                    <PasswordInput
                        size="md"
                        withAsterisk
                        label="Пароль"
                        placeholder="Введите пароль"
                        key={form.key('password')}
                        {...form.getInputProps('password')}
                    />

                    <Button
                        size="md"
                        type="submit"
                        fullWidth
                        mt="sm"
                        loading={submitting}
                    >
                        Войти
                    </Button>
                </Stack>
            </form>
        </Paper>
    )
}
