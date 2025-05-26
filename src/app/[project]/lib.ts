import fs from 'fs'
import path from 'path'

import { z } from 'zod'
import { formatDate, parseDateString } from '../lib'

const ProjectConfigSchema = z.object({
  id: z.string(),
  url: z.string().nullish(),
  date: z.union([z.string(), z.number()]).nullish(),
  name: z
    .object({
      en: z.string().nullish(),
      fi: z.string().nullish()
    })
    .nullish()
    .default({}),
  description: z
    .object({
      en: z.string().nullish(),
      fi: z.string().nullish()
    })
    .nullish()
    .default({}),
  icon: z.string().nullish(),
  largeImage: z.string().nullish(),
  images: z.array(z.string()).nullish(),
  pretext: z.string().nullish(),
  github: z.string().nullish(),
  ignoreDate: z
    .boolean()
    .nullish()
    .transform((v) => v ?? false),
  ignoreInListing: z
    .boolean()
    .nullish()
    .transform((v) => v ?? false),
  notAuthor: z
    .boolean()
    .nullish()
    .transform((v) => v ?? false)
})

type ProjectConfig = z.infer<typeof ProjectConfigSchema>

export type Project = {
  id: string
  name: string
  description?: string
  pretext?: string
  url?: string
  date?: number
  icon?: string
  images: string[]
  largeImage?: string
  github?: string
  notAuthor: boolean
  ignoreDate: boolean
  ignoreInListing: boolean
}

const getDirectories = (dir: string) => {
  return fs
    .readdirSync(dir)
    .filter((file) => fs.statSync(dir + '/' + file).isDirectory())
}

const emptyConfig = (id: string): ProjectConfig => {
  return {
    id,
    images: [],
    notAuthor: false,
    ignoreDate: true,
    ignoreInListing: true,
    name: {},
    description: {}
  }
}

const readConfig = (id: string): ProjectConfig => {
  try {
    const file = fs.readFileSync(
      path.join(process.cwd(), 'public', 'sub', id, 'project-config.json'),
      'utf-8'
    )
    const input = JSON.parse(file) as Record<string, unknown>

    return ProjectConfigSchema.parse(Object.assign(input, { id }))
  } catch {
    return emptyConfig(id)
  }
}

export const parseConfig = (config: ProjectConfig): Project => {
  return {
    id: config.id,
    url: config.url || undefined,
    date: parseDateString(config.date)?.getTime() || undefined,
    ignoreDate: config.ignoreDate,
    name: config.name?.en || config.name?.fi || config.id,
    description: config.description?.en || config.description?.fi || undefined,
    pretext: config.pretext || undefined,
    icon: config.icon || undefined,
    images: config.images || [],
    largeImage: config.largeImage || undefined,
    ignoreInListing: config.ignoreInListing,
    github: config.github || undefined,
    notAuthor: config.notAuthor
  }
}

export function getProjects() {
  return getDirectories(path.join(process.cwd(), 'public', 'sub'))
    .map(readConfig)
    .map(parseConfig)
    .sort((a, b) => {
      if (!a.date) return 1
      if (!b.date) return -1
      return b.date - a.date
    })
}

export function getFeaturedProjects() {
  return getProjects().filter((p) => !p.ignoreInListing)
}

export function getProjectById(id: string) {
  const config = readConfig(id)
  if (!config) return null
  return parseConfig(config)
}

export const groupEntries = <T>(
  projects: T[],
  getDate: (x: T) => Parameters<typeof formatDate>[0]
) =>
  Object.entries(
    Object.groupBy(
      projects,
      (project) =>
        formatDate(getDate(project))?.split(' ').at(-1) || 'Undated (2015-2018)'
    )
  ).sort((a, b) => Number(b[0]) - Number(a[0]))
