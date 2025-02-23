import { getProjectById, getProjects } from './lib'

export default async function Page({
  params
}: {
  params: Promise<{ project: string }>
}) {
  const id = (await params).project
  const project = getProjectById(id)!

  return (
    <div>
      <h1>{project.name}</h1>
    </div>
  )
}

export function generateStaticParams() {
  return getProjects().map(({ id }) => ({ project: id }))
}

export const dynamicParams = false
