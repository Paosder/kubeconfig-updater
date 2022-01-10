import { flow, makeObservable, observable } from 'mobx'
import { container, singleton } from 'tsyringe'
import dayjs, { Dayjs } from 'dayjs'
import { AggregatedClusterMetadata } from '../protos/kubeconfig_service_pb'
import browserLogger from '../logger/browserLogger'
import ClusterRepository from '../repositories/clusterRepository'

@singleton()
export default class ClusterMetadataStore {
  private readonly logger = browserLogger

  @observable
  state: 'ready' | 'fetch' | 'in-sync' = 'ready'

  @observable
  private _items: AggregatedClusterMetadata[] = []

  get items() {
    return this._items
  }

  // minute
  @observable
  ResyncInterval = 5

  private readonly syncedString = 'lastSynced'

  @observable
  private _lastSynced: Dayjs | null = dayjs(localStorage.getItem(this.syncedString))

  get lastSynced(): Dayjs | null {
    return this._lastSynced
  }

  set lastSynced(time: Dayjs | null) {
    this._lastSynced = time
    localStorage.setItem(this.syncedString, (time ?? dayjs('1970-01-01')).toISOString())
  }

  get shouldResync(): boolean {
    if (this.lastSynced === null) {
      return true
    }

    if (dayjs().diff(this.lastSynced, 'minute') >= this.ResyncInterval) {
      return true
    }

    return false
  }

  constructor() {
    makeObservable(this)
  }

  fetchMetadata = flow(function* (this: ClusterMetadataStore, resync?: boolean) {
    this._items = []

    if (resync || this.shouldResync) {
      this.logger.debug('request backend cluster metadata sync')
      this.state = 'in-sync'

      try {
        yield ClusterMetadataStore.sync()
      } catch (e) {
        this.logger.error(e)
      }
    }

    this.logger.debug('request backend cluster metadata fetch')
    this.state = 'fetch'

    try {
      this._items = yield ClusterMetadataStore.fetch()
    } catch (e) {
      this.logger.error(e)
    }

    this.logger.debug('fetch cluster metadata done.')
    this.state = 'ready'
  })

  private static async sync() {
    const repo = container.resolve(ClusterRepository)

    return repo.SyncAvailableClusters()
  }

  private static async fetch() {
    const repository = container.resolve(ClusterRepository)
    const res = await repository.GetAvailableClusters()

    return res.getClustersList()
  }
}