import { Graph, Thing, WithContext } from 'schema-dts'

export const JsonLd = ({ object }: { object: WithContext<Thing> | Graph }) => {
  return (
    <script
      type="application/ld+json"
      dangerouslySetInnerHTML={{ __html: JSON.stringify(object) }}
    />
  )
}
