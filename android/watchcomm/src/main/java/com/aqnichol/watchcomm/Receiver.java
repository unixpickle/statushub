package com.aqnichol.watchcomm;

import android.content.Context;
import android.os.Bundle;
import android.util.Log;

import com.google.android.gms.common.api.GoogleApiClient;
import com.google.android.gms.wearable.MessageApi;
import com.google.android.gms.wearable.MessageEvent;
import com.google.android.gms.wearable.Wearable;

import java.util.ArrayList;

/**
 * Receiver provides a synchronous API for polling for
 * messages from an external device.
 */
public class Receiver {
    private final ArrayList<MessageEvent> queue = new ArrayList<MessageEvent>();
    private Context context;
    private GoogleApiClient client;
    private MessageApi.MessageListener listener;

    /**
     * Create a new Receiver.
     *
     * @param ctx The context to use for google APIs.
     */
    public Receiver(Context ctx) {
        context = ctx;
        listener = new MessageApi.MessageListener() {
            @Override
            public void onMessageReceived(MessageEvent messageEvent) {
                pushMessage(messageEvent);
            }
        };
    }

    /**
     * Connect the receiver asynchronously.
     */
    public void connect() {
        if (client == null) {
            client = new GoogleApiClient.Builder(context)
                    .addApi(Wearable.API)
                    .addConnectionCallbacks(new GoogleApiClient.ConnectionCallbacks() {
                        @Override
                        public void onConnected(Bundle bundle) {
                            Wearable.MessageApi.addListener(client, listener);
                        }

                        @Override
                        public void onConnectionSuspended(int i) {
                            Wearable.MessageApi.removeListener(client, listener);
                        }
                    })
                    .build();
        }
        client.connect();
    }

    /**
     * Stop the Receiver until the next connect().
     */
    public void disconnect() {
        if (client.isConnected()) {
            Wearable.MessageApi.removeListener(client, listener);
            client.disconnect();
        }
    }

    /**
     * Remove all pending messages from the incoming message
     * queue, so that poll() will only see messages which
     * were received after clearQueue() was called.
     */
    public void clearQueue() {
        synchronized (queue) {
            queue.clear();
        }
    }

    /**
     * Poll receives the next message.
     *
     * @param timeoutMs The timeout in milliseconds.
     * @return The received message or null if the timeout elapsed.
     * @throws CommException If not connected to Play Services.
     */
    public MessageEvent poll(long timeoutMs) throws CommException {
        long startTime = System.currentTimeMillis();
        while (System.currentTimeMillis() < startTime + timeoutMs) {
            if (!client.isConnected()) {
                throw new CommException("not connected to Play Services");
            }
            synchronized (queue) {
                if (!queue.isEmpty()) {
                    MessageEvent res = queue.get(0);
                    queue.remove(0);
                    return res;
                }
                long remaining = timeoutMs + startTime - System.currentTimeMillis();
                if (remaining < 0) {
                    return null;
                }
                try {
                    queue.wait(remaining);
                } catch (InterruptedException e1) {
                }
            }
        }
        return null;
    }

    private void pushMessage(MessageEvent m) {
        synchronized (queue) {
            queue.add(m);
            queue.notifyAll();
        }
    }

}
