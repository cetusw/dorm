import { useEffect, useMemo, useState } from 'react'

import { UserIcon } from '@phosphor-icons/react'
import {
    Combobox,
    InputBase,
    Loader,
    Text,
    useCombobox,
} from '@mantine/core'

import classes from './ResidentSearchCombobox.module.css'

export type ResidentSearchOption = {
    id: string
    name: string
    description?: string | null
}

type Props = {
    label?: string
    withAsterisk?: boolean
    disabled?: boolean
    hideDropdownWhenSelected?: boolean
    selectedLabel?: string | null
    searchValue: string
    options: ResidentSearchOption[]
    loading: boolean
    error?: string | null
    placeholder: string
    selectedId: string | null
    onSearchChange: (value: string) => void
    onOptionSelect: (option: ResidentSearchOption) => void
    onFocus?: () => void
}

export function ResidentSearchCombobox({
    label,
    withAsterisk = false,
    disabled = false,
    hideDropdownWhenSelected = false,
    selectedLabel = null,
    searchValue,
    options,
    loading,
    error = null,
    placeholder,
    selectedId,
    onSearchChange,
    onOptionSelect,
    onFocus,
}: Props) {
    const combobox = useCombobox()
    const [focused, setFocused] = useState(false)
    const hasCommittedSelection = hideDropdownWhenSelected
        && selectedId !== null
        && selectedLabel != null
        && searchValue.trim() === selectedLabel.trim()

    useEffect(() => {
        if (!focused) {
            combobox.closeDropdown()
            return
        }

        if (disabled || hasCommittedSelection) {
            combobox.closeDropdown()
            return
        }

        combobox.openDropdown()
    }, [combobox, disabled, focused, hasCommittedSelection])

    const normalizedOptions = useMemo(
        () => options.map((option) => (
            <Combobox.Option key={option.id} value={option.id}>
                <div className={classes.option}>
                    <Text className={classes.optionName}>{option.name}</Text>
                    {option.description ? (
                        <Text className={classes.optionDescription}>{option.description}</Text>
                    ) : null}
                </div>
            </Combobox.Option>
        )),
        [options],
    )
    const leftSection = loading
        ? <Loader size={16} color="var(--app-color-text-muted)" />
        : <UserIcon size={18} weight="regular" />

    return (
        <Combobox
            store={combobox}
            onOptionSubmit={(value) => {
                const option = options.find((item) => item.id === value)
                if (!option) {
                    return
                }

                onOptionSelect(option)
                combobox.closeDropdown()
            }}
            withinPortal={false}
        >
            <Combobox.Target>
                <InputBase
                    label={label}
                    withAsterisk={withAsterisk}
                    disabled={disabled}
                    value={searchValue}
                    onChange={(event) => {
                        onSearchChange(event.currentTarget.value)
                        if (!disabled) {
                            combobox.openDropdown()
                        }
                    }}
                    onFocus={() => {
                        if (disabled) {
                            return
                        }

                        setFocused(true)
                        onFocus?.()
                    }}
                    onBlur={() => {
                        setFocused(false)
                        combobox.closeDropdown()
                    }}
                    placeholder={placeholder}
                    error={error}
                    leftSection={<span className={classes.leftSection}>{leftSection}</span>}
                    leftSectionPointerEvents="none"
                    rightSectionPointerEvents="none"
                    aria-autocomplete="list"
                    autoComplete="off"
                    data-selected-id={selectedId ?? undefined}
                />
            </Combobox.Target>

            <Combobox.Dropdown>
                <Combobox.Options>
                    {normalizedOptions.length > 0 ? (
                        normalizedOptions
                    ) : (
                        <Combobox.Empty>
                            {loading ? 'Загрузка...' : 'Ничего не найдено'}
                        </Combobox.Empty>
                    )}
                </Combobox.Options>
            </Combobox.Dropdown>
        </Combobox>
    )
}
