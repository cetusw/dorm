import type { IndividualTask } from '../model/types'
import { IndividualTaskResidentCard } from './IndividualTaskResidentCard'
import classes from './MyIndividualTasks.module.css'

type Props = { tasks: IndividualTask[]; pendingTaskID: string | null; onToggle: (task: IndividualTask) => void }

export function MyIndividualTasks({ tasks, pendingTaskID, onToggle }: Props) {
    const activeTasks = tasks.filter((task) => task.status !== 'verified')
    if (activeTasks.length === 0) return null

    return <section className={classes.section}>
        <h2 className={classes.heading}>Индивидуальные задачи</h2>
        {activeTasks.map((task) => {
            return <IndividualTaskResidentCard key={task.id} task={task} pending={pendingTaskID === task.id} onToggle={onToggle} />
        })}
    </section>
}
