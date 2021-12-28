export interface Project {
    id: string
    url: string | null
    date: number | null
    ignoreDate: boolean
    name: string
    description: string | null
    icon: string | null
}

export interface ProjectConfig {
    url: string | null
    date: string | number | null
    name: {
        en: string | null
        fi: string | null
    }
    description: {
        en: string | null
        fi: string | null
    }
    icon: string | null
    ignoreDate?: boolean
}
