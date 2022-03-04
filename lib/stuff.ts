import path from 'path'
import fs from 'fs'
import { Project, ProjectConfig } from '../types'
import { isNumber, isString } from './util'

const directory = path.join(process.cwd(), 'public/sub')

const getDirectories = (path: string) => {
    return fs.readdirSync(path).filter((file) => {
        return fs.statSync(path + '/' + file).isDirectory()
    })
}

export const getProjectById = (id: string): Project | null => {
    let data: ProjectConfig
    let date: number | null

    try {
        const file = fs.readFileSync(directory + '/' + id + '/project-config.json', 'utf-8')
        data = JSON.parse(file)
    } catch {
        return {
            id,
            url: null,
            date: null,
            ignoreDate: true,
            name: id,
            description: null,
            pretext: null,
            icon: null,
            images: [],
            ignoreInListing: true,
            github: null
        }
    }

    if (isString(data.date)) {
        const parts = data.date.split('.')
        if (!parts[2]) {
            date = null
        } else {
            date = new Date(parseInt(parts[2]), parseInt(parts[1]) - 1, parseInt(parts[0])).getTime()
        }
    } else if (isNumber(data.date)) {
        date = data.date
    } else {
        date = null
    }

    return {
        id,
        url: data?.url || null,
        date,
        ignoreDate: !!data?.ignoreDate,
        name: data?.name?.en || data?.name?.fi || id,
        description: data?.description?.en || data?.description?.fi || null,
        pretext: data.pretext || null,
        icon: data?.icon || null,
        images: data.images || [],
        ignoreInListing: !!data.ignoreInListing,
        github: data.github || null
    }
}

const getProjectNames = () => {
    return getDirectories(directory)
}

export const getProjectIds = () => {
    const names = getProjectNames()
    return names.map((name) => {
        return { params: { project: name } }
    })
}

export const getProjects = (): Project[] => {
    const directoryNames = getDirectories(directory)
    return directoryNames
        .map((id) => getProjectById(id))
        .filter(Boolean)
        .sort((a, b) => {
            a = a as Project
            b = b as Project
            if (!a.date || a.ignoreDate) return 1
            if (!b.date || b.ignoreDate) return -1
            return b.date - a.date
        }) as Project[]
}

export const getFilteredProjects = (): Project[] => {
    return getProjects().filter((p) => !p.ignoreInListing)
}
