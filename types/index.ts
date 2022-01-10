export interface Project {
    id: string
    url: string | null
    date: number | null
    ignoreDate: boolean
    ignoreInListing: boolean
    name: string
    description: string | null
    pretext: string | null
    icon: string | null
    images: string[]
    github: string | null
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
    ignoreInListing?: boolean
    ignoreDate?: boolean
    images?: string[]
    pretext?: string
    github?: string
}
