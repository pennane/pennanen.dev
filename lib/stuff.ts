import path from 'path'
import fs from 'fs'
import { TProject, TProjectConfig } from '../types/models'
import { isNumber, isString } from './util'

const directory = path.join(process.cwd(), 'public/sub')

const getDirectories = (path: string) => {
	return fs.readdirSync(path).filter((file) => {
		return fs.statSync(path + '/' + file).isDirectory()
	})
}

export const getProjectById = (id: string): TProject | null => {
	let data: TProjectConfig
	let date: number | null

	try {
		const file = fs.readFileSync(
			directory + '/' + id + '/project-config.json',
			'utf-8'
		)
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
			largeImage: null,
			ignoreInListing: true,
			github: null,
			notAuthor: null,
		}
	}

	if (isString(data.date)) {
		const parts = data.date.split('.')
		if (!parts[2]) {
			date = null
		} else {
			date = new Date(
				parseInt(parts[2]),
				parseInt(parts[1]) - 1,
				parseInt(parts[0])
			).getTime()
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
		largeImage: data.largeImage || null,
		ignoreInListing: !!data.ignoreInListing,
		github: data.github || null,
		notAuthor: data.notAuthor || null,
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

export const getProjects = (): TProject[] => {
	const directoryNames = getDirectories(directory)
	return directoryNames
		.map((id) => getProjectById(id))
		.filter(Boolean)
		.sort((a, b) => {
			a = a as TProject
			b = b as TProject
			if (!a.date || a.ignoreDate) return 1
			if (!b.date || b.ignoreDate) return -1
			return b.date - a.date
		}) as TProject[]
}

export const getFilteredProjects = (): TProject[] => {
	return getProjects().filter((p) => !p.ignoreInListing)
}
