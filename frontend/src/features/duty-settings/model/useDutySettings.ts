import { useCallback, useEffect, useState } from 'react'

import { ApiError } from '../../../shared/api/ApiError'
import { getDutySettings } from '../api/dutySettingsApi'
import type { DutySettingsArea, DutySettingsResponse, DutySettingsTask } from './types'

function sortAreas(areas: DutySettingsArea[]): DutySettingsArea[] {
    return [...areas].sort((left, right) => {
        if (left.floor == null && right.floor != null) {
            return 1
        }
        if (left.floor != null && right.floor == null) {
            return -1
        }
        if (left.floor != null && right.floor != null && left.floor !== right.floor) {
            return right.floor - left.floor
        }
        return left.name.localeCompare(right.name, 'ru')
    })
}

function sortTasks(tasks: DutySettingsTask[]): DutySettingsTask[] {
    return [...tasks].sort((left, right) => {
        if (left.is_included !== right.is_included) {
            return left.is_included ? -1 : 1
        }
        if (left.cost !== right.cost) {
            return right.cost - left.cost
        }
        return left.title.localeCompare(right.title, 'ru')
    })
}

function toError(error: unknown): string {
    if (error instanceof Error) {
        return error.message
    }

    return 'Не удалось загрузить настройки дежурства'
}

export function useDutySettings(groupId: string) {
    const [data, setData] = useState<DutySettingsResponse | null>(null)
    const [loading, setLoading] = useState(true)
    const [error, setError] = useState<string | null>(null)
    const [forbidden, setForbidden] = useState(false)

    const reload = useCallback(async (options?: { silent?: boolean }) => {
        if (!options?.silent) {
            setLoading(true)
        }
        setError(null)
        setForbidden(false)

        try {
            const response = await getDutySettings(groupId)
            setData({
                ...response,
                areas: sortAreas(response.areas.map((area) => ({
                    ...area,
                    tasks: sortTasks(area.tasks),
                }))),
            })
        } catch (error) {
            if (error instanceof ApiError && error.status === 403) {
                setForbidden(true)
                setError(error.message)
                setData(null)
            } else {
                setError(toError(error))
            }
        } finally {
            if (!options?.silent) {
                setLoading(false)
            }
        }
    }, [groupId])

    useEffect(() => {
        void reload()
    }, [reload])

    return {
        data,
        loading,
        error,
        forbidden,
        reload,
        appendArea: useCallback((area: DutySettingsArea) => {
            setData((current) => {
                if (!current) {
                    return current
                }

                return {
                    ...current,
                    areas: sortAreas([...current.areas, area]),
                }
            })
        }, []),
        replaceArea: useCallback((area: DutySettingsArea) => {
            setData((current) => {
                if (!current) {
                    return current
                }

                return {
                    ...current,
                    areas: sortAreas(current.areas.map((item) => (item.id === area.id ? area : item))),
                }
            })
        }, []),
        removeArea: useCallback((areaId: number) => {
            setData((current) => {
                if (!current) {
                    return current
                }

                return {
                    ...current,
                    areas: current.areas.filter((area) => area.id !== areaId),
                }
            })
        }, []),
        appendTask: useCallback((areaId: number, task: DutySettingsTask) => {
            setData((current) => {
                if (!current) {
                    return current
                }

                return {
                    ...current,
                    areas: current.areas.map((area) => area.id === areaId
                        ? {
                            ...area,
                            tasks: sortTasks([...area.tasks, task]),
                        }
                        : area),
                }
            })
        }, []),
        replaceTask: useCallback((taskId: string, updatedTask: DutySettingsTask, nextAreaId: number) => {
            setData((current) => {
                if (!current) {
                    return current
                }

                return {
                    ...current,
                    areas: sortAreas(current.areas.map((area) => {
                        const nextTasks = area.tasks.filter((task) => task.id !== taskId)
                        if (area.id !== nextAreaId) {
                            return { ...area, tasks: sortTasks(nextTasks) }
                        }

                        return {
                            ...area,
                            tasks: sortTasks([...nextTasks, updatedTask]),
                        }
                    })),
                }
            })
        }, []),
        updateTask: useCallback((taskId: string, updatedTask: DutySettingsTask, nextAreaId: number) => {
            setData((current) => {
                if (!current) {
                    return current
                }

                const areasWithoutTask = current.areas.map((area) => ({
                    ...area,
                    tasks: area.tasks.filter((task) => task.id !== taskId),
                }))

                return {
                    ...current,
                    areas: areasWithoutTask.map((area) => area.id === nextAreaId
                        ? {
                            ...area,
                            tasks: sortTasks([...area.tasks, updatedTask]),
                        }
                        : area),
                }
            })
        }, []),
        removeTask: useCallback((taskId: string) => {
            setData((current) => {
                if (!current) {
                    return current
                }

                return {
                    ...current,
                    areas: current.areas.map((area) => ({
                        ...area,
                        tasks: area.tasks.filter((task) => task.id !== taskId),
                    })),
                }
            })
        }, []),
    }
}
