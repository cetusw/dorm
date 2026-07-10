import {Button, Paper, PasswordInput, Stack, TextInput, Title} from '@mantine/core'
import {useForm} from '@mantine/form'

import './LoginPage.css'

type LoginFormValues = {
    login: string
    password: string
}

export function LoginPage() {
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
        <div className="loginPage">
            <Paper className="loginCard" radius="lg" p="xl" shadow="sm" withBorder>
                <Title order={2} ta="center">
                    Добро пожаловать
                </Title>

                <form
                    onSubmit={form.onSubmit(() => {
                        return
                    })}
                >
                    <Stack gap="md" mt="xl">
                        <TextInput
                            withAsterisk
                            label="Логин"
                            placeholder="Введите логин"
                            key={form.key('login')}
                            {...form.getInputProps('login')}
                        />

                        <PasswordInput
                            withAsterisk
                            label="Пароль"
                            placeholder="Введите пароль"
                            key={form.key('password')}
                            {...form.getInputProps('password')}
                        />

                        <Button type="submit" fullWidth mt="sm">
                            Войти
                        </Button>
                    </Stack>
                </form>
            </Paper>
        </div>
    )
}
