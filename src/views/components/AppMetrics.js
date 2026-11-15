export default {
    name: 'AppMetrics',
    data() {
        return {
            loading: false,
            metrics: null,
            error: '',
            refreshedAt: null,
        }
    },
    computed: {
        results() {
            return this.metrics?.results || {}
        },
        system() {
            return this.results.system || {}
        },
        devices() {
            return this.results.devices || { total: 0, logged_in: 0, connected: 0 }
        },
        memory() {
            return this.system.memory || {}
        },
        folders() {
            return this.system.folders || {}
        },
        filesystem() {
            return this.system.filesystem || {}
        },
    },
    methods: {
        openModal() {
            $('#modalAppMetrics')
                .modal({
                    autofocus: false,
                    observeChanges: true,
                })
                .modal('show')
            this.loadMetrics()
        },
        async loadMetrics() {
            if (this.loading) return
            this.loading = true
            this.error = ''
            try {
                const response = await window.http.get('/app/metrics')
                this.metrics = response.data || null
                this.refreshedAt = new Date()
            } catch (error) {
                this.error = error.response?.data?.message || error.message || 'Failed to fetch system metrics'
                showErrorInfo(this.error)
            } finally {
                this.loading = false
            }
        },
        formatBytes(value) {
            if (value === null || value === undefined || Number.isNaN(Number(value))) return 'Unknown'
            const bytes = Math.max(0, Number(value))
            const units = ['B', 'KB', 'MB', 'GB', 'TB']
            let size = bytes
            let unit = 0
            while (size >= 1024 && unit < units.length - 1) {
                size /= 1024
                unit++
            }
            const digits = unit === 0 ? 0 : (size >= 100 ? 0 : 1)
            return `${size.toFixed(digits)} ${units[unit]}`
        },
        formatDuration(value) {
            const total = Math.max(0, Math.floor(Number(value) || 0))
            const days = Math.floor(total / 86400)
            const hours = Math.floor((total % 86400) / 3600)
            const minutes = Math.floor((total % 3600) / 60)
            const seconds = total % 60
            const parts = []
            if (days) parts.push(`${days}d`)
            if (hours || days) parts.push(`${hours}h`)
            if (minutes || hours || days) parts.push(`${minutes}m`)
            parts.push(`${seconds}s`)
            return parts.join(' ')
        },
        formatNumber(value, digits = 0) {
            if (value === null || value === undefined || Number.isNaN(Number(value))) return 'Unknown'
            return Number(value).toLocaleString(undefined, { maximumFractionDigits: digits })
        },
        formatDate(value) {
            if (!value) return 'Unknown'
            const date = new Date(value)
            if (Number.isNaN(date.getTime())) return String(value)
            return date.toLocaleString()
        },
    },
    template: `
        <div class="teal card" @click="openModal" style="cursor: pointer">
            <div class="content">
                <a class="ui blue right ribbon label">System</a>
                <div class="header">System Metrics</div>
                <div class="description">View API uptime, devices, memory, storage and disk capacity.</div>
            </div>
        </div>

        <div class="ui large modal metrics-modal" id="modalAppMetrics">
            <i class="close icon"></i>
            <div class="header metrics-modal-header">
                <i class="chart line icon"></i>
                System Metrics
            </div>

            <div class="scrolling content metrics-modal-content">
                <div v-if="loading && !metrics" class="metrics-loading">
                    <div class="ui active centered inline loader"></div>
                    <span>Loading system information...</span>
                </div>

                <div v-if="error" class="metrics-error">
                    <strong>Metrics unavailable</strong>
                    <span>{{ error }}</span>
                </div>

                <template v-if="metrics">
                    <div class="metrics-overview">
                        <section class="metrics-panel">
                            <div class="metrics-panel-title"><i class="server icon"></i> API</div>
                            <div class="metrics-row"><span>Version</span><strong>{{ results.version || 'Unknown' }}</strong></div>
                            <div class="metrics-row"><span>Uptime</span><strong>{{ formatDuration(system.uptime_seconds) }}</strong></div>
                            <div class="metrics-row"><span>Started</span><strong>{{ formatDate(system.started_at) }}</strong></div>
                            <div class="metrics-row"><span>Go version</span><strong>{{ system.go_version || 'Unknown' }}</strong></div>
                        </section>

                        <section class="metrics-panel">
                            <div class="metrics-panel-title"><i class="mobile alternate icon"></i> Devices</div>
                            <div class="metrics-device-grid">
                                <div><strong>{{ devices.total || 0 }}</strong><span>Total</span></div>
                                <div><strong>{{ devices.logged_in || 0 }}</strong><span>Logged in</span></div>
                                <div><strong>{{ devices.connected || 0 }}</strong><span>Connected</span></div>
                            </div>
                        </section>

                        <section class="metrics-panel">
                            <div class="metrics-panel-title"><i class="database icon"></i> Storage mode</div>
                            <div class="metrics-row"><span>Session store</span><strong>{{ results.session_store || 'Unknown' }}</strong></div>
                            <div class="metrics-row">
                                <span>Chat storage</span>
                                <strong class="metrics-status" :class="results.chat_storage_persistent ? 'is-persistent' : 'is-temporary'">
                                    {{ results.chat_storage_persistent ? 'Persistent' : 'Temporary' }}
                                </strong>
                            </div>
                        </section>
                    </div>

                    <div class="metrics-section-title"><i class="memory icon"></i> Memory and runtime</div>
                    <div class="metrics-tiles">
                        <div class="metrics-tile"><span>Heap allocated</span><strong>{{ formatBytes(memory.heap_alloc_bytes) }}</strong></div>
                        <div class="metrics-tile"><span>Runtime reserved</span><strong>{{ formatBytes(memory.runtime_sys_bytes) }}</strong></div>
                        <div class="metrics-tile"><span>Process RSS</span><strong>{{ formatBytes(memory.process_rss_bytes) }}</strong></div>
                        <div class="metrics-tile"><span>Container usage</span><strong>{{ formatBytes(memory.container_bytes) }}</strong></div>
                        <div class="metrics-tile"><span>Container limit</span><strong>{{ formatBytes(memory.container_limit_bytes) }}</strong></div>
                        <div class="metrics-tile"><span>CPU time</span><strong>{{ formatNumber(system.cpu_seconds, 2) }} s</strong></div>
                        <div class="metrics-tile"><span>Goroutines</span><strong>{{ formatNumber(memory.goroutines) }}</strong></div>
                        <div class="metrics-tile"><span>GC cycles</span><strong>{{ formatNumber(memory.gc_cycles) }}</strong></div>
                    </div>

                    <div class="metrics-section-title"><i class="hdd icon"></i> Disk and folders</div>
                    <div class="metrics-storage-list">
                        <div class="metrics-storage-row">
                            <div><i class="images outline icon"></i><strong>Media</strong></div>
                            <span>{{ formatBytes(folders.media?.bytes) }}</span>
                            <small>{{ formatNumber(folders.media?.files) }} files</small>
                        </div>
                        <div class="metrics-storage-row">
                            <div><i class="folder outline icon"></i><strong>Storages</strong></div>
                            <span>{{ formatBytes(folders.storages?.bytes) }}</span>
                            <small>{{ formatNumber(folders.storages?.files) }} files</small>
                        </div>
                        <div class="metrics-storage-row">
                            <div><i class="hdd outline icon"></i><strong>Filesystem total</strong></div>
                            <span>{{ formatBytes(filesystem.total_bytes) }}</span>
                            <small>Capacity</small>
                        </div>
                        <div class="metrics-storage-row">
                            <div><i class="check circle outline icon"></i><strong>Filesystem free</strong></div>
                            <span>{{ formatBytes(filesystem.free_bytes) }}</span>
                            <small>Available</small>
                        </div>
                    </div>

                    <div v-if="refreshedAt" class="metrics-refreshed">
                        <i class="clock outline icon"></i>
                        Last refreshed: {{ refreshedAt.toLocaleString() }}
                    </div>
                </template>
            </div>

            <div class="actions metrics-actions">
                <button class="ui button" @click="loadMetrics" :class="{loading: loading}" :disabled="loading">
                    <i class="sync alternate icon"></i> Refresh
                </button>
                <button class="ui positive button"><i class="check icon"></i> Close</button>
            </div>
        </div>
    `
}
