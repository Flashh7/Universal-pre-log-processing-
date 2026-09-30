import Card from "../components/card";
import DetailGrid from "../components/detail_grid";
import type { LogEvent } from "../types/log";

interface EventDetailsProps {
  event: LogEvent;
}

function EventDetails({ event }: EventDetailsProps) {
  const overviewDetails = [
    { label: "Event ID", value: event.event_uid },
    { label: "Time", value: event.time },
    { label: "Severity", value: event.severity },
    { label: "Severity ID", value: event.severity_id },
    { label: "Status", value: event.status },
    { label: "Status ID", value: event.status_id },
  ];

  const classificationDetails = [
    { label: "Category", value: event.category_name },
    { label: "Category UID", value: event.category_uid },
    { label: "Class", value: event.class_name },
    { label: "Class UID", value: event.class_uid },
    { label: "Activity", value: event.activity_name },
    { label: "Activity ID", value: event.activity_id },
    { label: "Type", value: event.type_name },
    { label: "Type UID", value: event.type_uid },
  ];

  return (
    <div className="event-details">
      <Card title="Event Overview">
        <DetailGrid items={overviewDetails} />
      </Card>

      <Card title="Event Classification">
        <DetailGrid items={classificationDetails} />
      </Card>

      <Card title="Raw Event">
        <pre className="raw-event">
          {event.raw_event}
        </pre>
      </Card>
    </div>
  );
}

export default EventDetails;