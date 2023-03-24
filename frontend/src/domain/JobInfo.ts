export interface JobInfo {
    name: string,
    namespace: string,
    taskGroups: TaskGroupInfo[]
}
export interface TaskGroupInfo {
    name: string,
    status: string,
    allocations: AllocationInfo[]
}
export interface AllocationInfo {
    id: string,
    tasks: TaskInfo[]
}
export interface TaskInfo {
    name: string,
    status: string,
    startedAt: string,
    finishedAt: string,
    lastRestart: string
}
