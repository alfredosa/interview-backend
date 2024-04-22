# Assignment Presentation

# Table of Contents
- [Intro](#Intro)
- [Disclaimers](#Disclaimers)
- [Definitions](#definitions)
- [Demo](#Demo)
- [Part 1](#part-1-implement-backend-api)
- [Part 2](#part-two-store-delivery-units)
- [Main Topic](#main-topic)
- [Details](#details)

## Disclaimers:

- A lot of unknowns. E.g How many warehouses or Suppliers there are / Or the amount of traffic.
- Code Structure and Style agreements. I went with a hunch but I would agree with the team before. I tried to keep it close to the current state.


## Definitions

> **Warehouse**: A warehouse is a Store that receives Deliverables, delivered by Cargo Units. It has a Location.<br>
> **Cargo Unit**: A Supplier of Deliverables. Delivers into Warehouses. It has a Location.<br>
> **Location**: Longitude and Lattidue (Position on a cartesian map) of a given Object.<br>

## Demo 

<p align="center">
  <img src="/docs/assets/demo.png" alt="Warehouse and Supplier Models" width="50%">
</p>


### Part 1: Implement Backend API

> #### Technology Choice:
> I went with **gRPC**. For this use case, strongly typed messages or schema, it's streaming capabilities, its a strong candidate for vehicle and cargo tracking.

#### Server:

<div style="display: flex; justify-content: space-around; align-items: center;">
  <div style="flex: 1; padding: 10px;">
    <img src="/docs/assets/server.png" alt="Server" style="width: 100%;">
  </div>
  <div style="flex: 1; padding: 10px;">
    <img src="/docs/assets/grpcservice.png" alt="Service" style="width: 100%;">
  </div>
</div>


#### Handlers

<p align="center">
  <img src="/docs/assets/handlersc.png" alt="Handlers" width="100%">
</p>

<!-- #### Logging STDOUT Every Second

<p align="center">
  <img src="/docs/assets/printServerStats.png" alt="Warehouse and Supplier Models" width="50%">
</p> -->

<!-- ### Models:

<p align="center">
  <img src="/docs/assets/models.png" alt="Warehouse and Supplier Models" width="37%">
</p> -->

### Part Two: Store Delivery Units

1. Implement storage for these delivery paths, with incoming data processed by
   your API.

#### **The Controller :O**

<p align="center">
  <img src="/docs/assets/controller.png" alt="Warehouse and Supplier Models" width="100%">
</p>

<!-- 
<div style="display: flex; justify-content: space-around; align-items: center;">
  <div style="flex: 1; padding: 10px;">
    <img src="/docs/assets/warehouseunit.png" alt="Server" style="width: 100%;">
  </div>
  <div style="flex: 1; padding: 10px;">
    <img src="/docs/assets/moveunit.png" alt="Service" style="width: 100%;">
  </div>
</div> -->

2. Write at least one unit test to verify the functionality of your solution.
3. Your API Server (solution) must include the following outputs:

Extras:

- Periodic Logging: Every second, output a log message to STDOUT that displays
  the count of messages received during that second.


- Summary Output: Provide a summary that includes:
  - The total number of delivery units.
  - A list of warehouses that have received supplies (i.e., units that have
    reached their destination).
  - The quantity of delivery units each warehouse has received.