import { Center } from '@mantine/core'

import { LoginForm } from '../../features/auth/ui/LoginForm'

import classes from './LoginPage.module.css'

export function LoginPage() {
    return (
        <div className={classes.loginPage}>
            <Center mih="100dvh" px={{ base: 'md', sm: 'xl' }}>
                <LoginForm />
            </Center>
        </div>
    )
}
